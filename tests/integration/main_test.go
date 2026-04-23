//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/pkg/tel"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var gatewayURL string

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Resolve imposters dir relative to this package (go test sets cwd to package dir)
	impostersDir, err := filepath.Abs("../../imposters")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve imposters path: %v\n", err)
		os.Exit(1)
	}

	bankC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "bbyars/mountebank:2.8.1",
			ExposedPorts: []string{"2525/tcp", "8080/tcp"},
			Cmd:          []string{"--configfile", "/imposters/bank_simulator.ejs", "--allowInjection"},
			Mounts: testcontainers.Mounts(
				testcontainers.BindMount(impostersDir, "/imposters"),
			),
			// Wait until the mountebank management API confirms the imposter is loaded
			WaitingFor: wait.ForHTTP("/imposters").WithPort("2525/tcp"),
		},
		Started: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start bank container: %v\n", err)
		os.Exit(1)
	}
	defer bankC.Terminate(ctx) //nolint:errcheck

	bankHost, _ := bankC.Host(ctx)
	bankPort, _ := bankC.MappedPort(ctx, "8080/tcp")
	bankURL := fmt.Sprintf("http://%s:%s", bankHost, bankPort.Port())

	gw := api.New(bankURL, tel.NewNoopTelemetry())
	server := httptest.NewServer(gw.Handler())
	defer server.Close()

	gatewayURL = server.URL

	os.Exit(m.Run())
}
