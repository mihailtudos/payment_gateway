package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
)

var (
	instance Config
	once     sync.Once
	errParse error
)

type HTTP struct {
	ReadTimeout       int `env:"HTTP_READ_TIMEOUT"    envDefault:"3"`
	ReadHeaderTimeout int `env:"HTTP_READ_HEADER_TIMEOUT"    envDefault:"3"`
	WriteTimeout      int `env:"HTTP_WRITE_TIMEOUT"    envDefault:"10"`
	IdleTimeout       int `env:"HTTP_IDLE_TIMEOUT"    envDefault:"60"`
	Port              int `env:"HTTP_PORT"    envDefault:"8090"`
}
type Config struct {
	ServiceName         string `env:"SERVICE_NAME"    envDefault:"payment-gateway"`
	ServiceVersion      string `env:"SERVICE_VERSION" envDefault:"0.1.0"`
	BankURL             string `env:"ACQUIRING_BANK_URL"        envDefault:"http://localhost:8080"`
	OTLPEndpoint        string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"   envDefault:"localhost:4317"`
	OTelShutdownTimeout int    `env:"OTEL_SHUTDOWN_TIMEOUT"   envDefault:"5"`
	HTTP
}

func MustNewConfig(version string) Config {
	once.Do(func() {
		errParse = env.Parse(&instance)
	})
	if errParse != nil {
		panic(errParse)
	}
	if version != "" {
		instance.ServiceVersion = version
	}
	return instance
}
