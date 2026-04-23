package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/adapters/bank"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api/dto"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/httputil"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/cko-recruitment/payment-gateway-challenge-go/pkg/tel"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

type PaymentsHandler struct {
	storage       *repository.PaymentsRepository
	acquiringBank bank.Adapter
	telemetry     *tel.Telemetry
	getCounter    metric.Int64Counter
	postCounter   metric.Int64Counter
}

func NewPaymentsHandler(storage *repository.PaymentsRepository, bankAdapter bank.Adapter, t *tel.Telemetry) *PaymentsHandler {
	meter := t.Meter()

	getCounter, _ := meter.Int64Counter("payment.get.requests",
		metric.WithDescription("Total GET /payments/{id} requests by outcome"),
	)
	postCounter, _ := meter.Int64Counter("payment.post.requests",
		metric.WithDescription("Total POST /payments requests by outcome and currency"),
	)

	return &PaymentsHandler{
		storage:       storage,
		acquiringBank: bankAdapter,
		telemetry:     t,
		getCounter:    getCounter,
		postCounter:   postCounter,
	}
}

func (h *PaymentsHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		ctx, span := h.telemetry.TraceStart(r.Context(), "GetPayment")
		defer span.End()
		span.SetAttributes(attribute.String("payment.id", id))

		log := httputil.LoggerFromContext(ctx)

		pID, err := uuid.Parse(id)
		if err != nil || pID == uuid.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid payment id")
			log.WarnContext(ctx,
				"invalid payment id",
				slog.String("payment_id", id),
			)
			h.getCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "invalid_id")))
			httputil.BadRequest(w, "invalid payment id")
			return
		}

		_, storageSpan := h.telemetry.Tracer().Start(ctx, "storage.GetPayment")
		payment := h.storage.GetPayment(pID)
		storageSpan.End()

		if payment != nil {
			span.SetAttributes(
				attribute.String("payment.status",
					string(payment.PaymentStatus)),
			)
			log.InfoContext(ctx,
				"payment found",
				slog.String("payment_id", id),
				slog.String("status", string(payment.PaymentStatus)),
			)
			h.getCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "found")))
			httputil.OK(w, payment)
		} else {
			log.InfoContext(ctx,
				"payment not found",
				slog.String("payment_id", id),
			)
			h.getCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "not_found")))
			httputil.NotFound(w, "payment not found")
		}
	}
}

func (h *PaymentsHandler) PostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := h.telemetry.TraceStart(r.Context(), "PostPayment")
		defer span.End()

		log := httputil.LoggerFromContext(ctx)

		var body models.PostPaymentRequest
		if err := httputil.DecodeJSON(r.Body, &body); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "malformed request body")
			log.WarnContext(ctx, "malformed request body", slog.Any("err", err))
			h.postCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "bad_request")))
			httputil.BadRequest(w, "request body is malformed or missing")
			return
		}
		defer r.Body.Close()

		if err := body.Validate(); err != nil {
			span.SetStatus(codes.Error, "validation failed")
			log.WarnContext(ctx, "payment request validation failed", slog.Any("err", err))
			h.postCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "validation_failed")))
			httputil.ValidationFailed(w, httputil.TranslateValidationErrors(err))
			return
		}

		span.SetAttributes(
			attribute.String("payment.currency", body.Currency),
			attribute.Int64("payment.amount", int64(body.Amount)),
		)

		bpReq := dto.FromPaymentRequestToBankReq(body)

		bankCtx, bankSpan := h.telemetry.Tracer().Start(ctx, "bank.ProcessPayment")
		bankSpan.SetAttributes(
			attribute.String("payment.currency", body.Currency),
			attribute.Int64("payment.amount", int64(body.Amount)),
		)
		bpRes, err := h.acquiringBank.ProcessPayment(bankCtx, bpReq)
		if err != nil {
			bankSpan.RecordError(err)
			bankSpan.SetStatus(codes.Error, "bank processing failed")
			bankSpan.End()

			span.SetStatus(codes.Error, "bank processing failed")

			switch {
			case errors.Is(err, bank.ErrBankUnavailable):
				log.ErrorContext(ctx, "acquiring bank unavailable", slog.Any("err", err))
				h.postCounter.Add(ctx, 1, metric.WithAttributes(
					attribute.String("outcome", "bank_unavailable"),
					attribute.String("currency", body.Currency),
				))
				httputil.ServiceUnavailable(w, "payment processor is temporarily unavailable")
			default:
				log.ErrorContext(ctx, "acquiring bank error", slog.Any("err", err))
				h.postCounter.Add(ctx, 1, metric.WithAttributes(
					attribute.String("outcome", "bank_error"),
					attribute.String("currency", body.Currency),
				))
				httputil.InternalServerError(w)
			}
			return
		}
		bankSpan.SetAttributes(attribute.Bool("bank.authorized", bpRes.Authorized))
		bankSpan.End()

		payment := dto.FromBankResToPaymentResp(body, bpRes)

		_, storageSpan := h.telemetry.Tracer().Start(ctx, "storage.AddPayment")
		err = h.storage.AddPayment(payment)
		storageSpan.End()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "storage failed")
			log.ErrorContext(ctx, "failed to store payment", slog.Any("err", err))
			h.postCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("outcome", "storage_error"),
				attribute.String("currency", body.Currency),
			))
			httputil.InternalServerError(w)
			return
		}

		span.SetAttributes(
			attribute.String("payment.id", payment.ID),
			attribute.String("payment.status", string(payment.PaymentStatus)),
		)

		if payment.PaymentStatus == models.Declined {
			log.InfoContext(ctx, "payment declined", slog.String("payment_id", payment.ID), slog.String("currency", body.Currency))
			h.postCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.String("outcome", "declined"),
				attribute.String("currency", body.Currency),
			))
			httputil.PaymentDeclined(w, payment)
			return
		}

		log.InfoContext(ctx, "payment authorized", slog.String("payment_id", payment.ID), slog.String("currency", body.Currency), slog.Int("amount", body.Amount))
		h.postCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("outcome", "authorized"),
			attribute.String("currency", body.Currency),
		))
		httputil.Created(w, payment)
	}
}
