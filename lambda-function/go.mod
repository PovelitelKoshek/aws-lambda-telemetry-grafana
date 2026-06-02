module lambda-grafana-go

go 1.22

require (
	github.com/aws/aws-lambda-go v1.47.0
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.34.0
	go.opentelemetry.io/otel/metric v1.34.0
	go.opentelemetry.io/otel/sdk v1.34.0
	go.opentelemetry.io/otel/sdk/metric v1.34.0
	go.opentelemetry.io/otel/sdk/resource v1.34.0
	go.opentelemetry.io/otel/semconv/v1.26.0 v1.26.0
)
