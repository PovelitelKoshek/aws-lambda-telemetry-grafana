package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Event struct {
	Name string `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

var (
	meterProvider *sdkmetric.MeterProvider
	invocations   metric.Int64Counter
	errorsCount   metric.Int64Counter
	durationMs    metric.Float64Histogram
)

func initMeter(ctx context.Context) func(context.Context) error {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "go-lambda-demo"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("cloud.provider", "aws"),
			attribute.String("cloud.platform", "aws_lambda"),
		),
	)
	if err != nil {
		log.Printf("resource error: %v", err)
	}

	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		log.Fatalf("failed to create metric exporter: %v", err)
	}

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(2*time.Second),
			),
		),
	)

	otel.SetMeterProvider(meterProvider)

	meter := otel.Meter("lambda-grafana-go")

	invocations, err = meter.Int64Counter("lambda_invocations_total")
	if err != nil {
		log.Fatalf("counter error: %v", err)
	}

	errorsCount, err = meter.Int64Counter("lambda_errors_total")
	if err != nil {
		log.Fatalf("error counter error: %v", err)
	}

	durationMs, err = meter.Float64Histogram("lambda_duration_ms", metric.WithUnit("ms"))
	if err != nil {
		log.Fatalf("histogram error: %v", err)
	}

	return meterProvider.Shutdown
}

func handler(ctx context.Context, event Event) (Response, error) {
	start := time.Now()

	invocations.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("function", "Test2"),
		),
	)

	logLine := map[string]interface{}{
		"level":    "info",
		"service":  "go-lambda-demo",
		"message":  "lambda invocation started",
		"event":    event,
		"datetime": time.Now().Format(time.RFC3339),
	}

	b, _ := json.Marshal(logLine)
	log.Println(string(b))

	if event.Name == "error" {
		errorsCount.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("function", "Test2"),
			),
		)

		duration := float64(time.Since(start).Milliseconds())
		durationMs.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("function", "Test2"),
			),
		)

		return Response{
			Message: "simulated error metric was recorded",
		}, nil
	}

	duration := float64(time.Since(start).Milliseconds())
	durationMs.Record(ctx, duration,
		metric.WithAttributes(
			attribute.String("function", "Test2"),
		),
	)

	return Response{
		Message: "Hello, " + event.Name,
	}, nil
}

func main() {
	ctx := context.Background()

	shutdown := initMeter(ctx)
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	lambda.Start(handler)
}
