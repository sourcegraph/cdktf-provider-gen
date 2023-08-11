package resource

import (
	"context"
	"os"

	"github.com/sourcegraph/log"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func BuildLogResource(name string, buildCommit string) log.Resource {
	return log.Resource{
		Name:    name,
		Version: buildCommit,
		InstanceID: func() string {
			if envHostname := os.Getenv("HOSTNAME"); envHostname != "" {
				return envHostname
			}
			h, _ := os.Hostname()
			return h
		}(),
	}
}

func BuildOpenTelemetryResource(ctx context.Context, r log.Resource, opts ...otelresource.Option) (*otelresource.Resource, error) {
	opts = append(opts,
		// Add your own custom attributes to identify your application
		// Do not provide a schema URL here so that we can use detectors
		otelresource.WithAttributes(
			semconv.ServiceNameKey.String(r.Name),
			semconv.ServiceNamespaceKey.String(r.Namespace),
			semconv.ServiceInstanceIDKey.String(r.InstanceID),
			semconv.ServiceVersionKey.String(r.Version)))
	return otelresource.New(ctx, opts...)
}
