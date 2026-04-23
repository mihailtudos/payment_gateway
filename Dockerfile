FROM golang:1.26.2-alpine3.22 AS builder
RUN apk update && apk add --no-cache openssh-client git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go generate ./...
# Define Build Arguments
ARG VERSION=1.0.0
ARG COMMIT_HASH
ARG BUILD_DATE

RUN go build -ldflags "\
    -X main.version=${VERSION} \
    -X main.commit=${COMMIT_HASH} \
    -X main.date=${BUILD_DATE}" \
    -o pgw ./cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/pgw .
# Set the binary as the entrypoint
ENTRYPOINT ["./pgw"]