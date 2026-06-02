# AWS Lambda Metrics to Grafana Cloud through Grafana Alloy

This project demonstrates how to send custom telemetry from an AWS Lambda function written in Go to Grafana Cloud using OpenTelemetry, OTLP/HTTP and Grafana Alloy.

The main architecture is:

```
AWS Lambda on Go
   ↓ OTLP / HTTP
EC2 Tier with Grafana Alloy
   ↓
Grafana Cloud
   ↓
Grafana Dashboard
```

# Project goal

The goal of this project was to send AWS Lambda telemetry to Grafana without using CloudWatch as the main Grafana data source.

Instead of reading Lambda metrics from CloudWatch, the Lambda function sends OpenTelemetry metrics directly to a Grafana Alloy collector running on an EC2 instance. Alloy then forwards the data to Grafana Cloud.


# In this project:

a simple AWS Lambda function was written in Go;
OpenTelemetry SDK was added to the Lambda code;
custom Lambda metrics were created in the Go code;
an EC2 instance was created as a telemetry collector;
Grafana Alloy was installed on EC2;
Alloy was configured to receive OTLP/HTTP data on port 4318;
Alloy forwarded received telemetry to Grafana Cloud;
Grafana dashboards were created to visualize Lambda metrics.

# Final architecture

AWS Lambda Test2
        ↓
OpenTelemetry SDK in Go
        ↓
OTLP/HTTP request to EC2 public IP:4318
        ↓
Grafana Alloy on EC2
        ↓
Grafana Cloud Metrics
        ↓
Grafana Dashboard

# AWS Lambda

The Lambda function was written in Go. The function receives a test event, writes a simple log message and records custom OpenTelemetry metrics.

The function was deployed using custom runtime:

Runtime: provided.al2023
Handler: bootstrap
Architecture: x86_64

The Go code was compiled into a Linux binary named bootstrap and uploaded to AWS Lambda as a zip archive.

Because bootstrap is a compiled Go binary, AWS Lambda Console cannot open it as a text file. This is normal.

OpenTelemetry metrics

The Lambda function creates and sends the following custom metrics:

```
lambda_invocations_total
lambda_errors_total
lambda_duration_ms
```

These metrics are created inside the Go Lambda code using OpenTelemetry SDK.

The Lambda uses environment variables to know where to send OTLP data:

```
OTEL_EXPORTER_OTLP_ENDPOINT=http://13.60.40.159:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
OTEL_SERVICE_NAME=go-lambda-demo
OTEL_RESOURCE_ATTRIBUTES=service.namespace=aws-lambda,deployment.environment=production
```

The most important variable is:

```
OTEL_EXPORTER_OTLP_ENDPOINT=http://X.X.X.X:4318
```

This value points to the EC2 instance where Grafana Alloy is running.

## EC2 instance

A separate EC2 instance was created to run Grafana Alloy.

EC2 configuration used in this project:

```
Name: grafana-alloy-collector
AMI: Ubuntu Server 24.04 LTS
Instance type: t3.micro
Region: eu-north-1
Architecture: x86_64
Storage: 8 GiB gp3
Public IPv4: X.X.X.X
```

The public IPv4 address of the EC2 instance was used as the OTLP endpoint for the Lambda function.

## Security Group

The EC2 Security Group was configured with inbound rules:

```
SSH          TCP 22    My IP
Custom TCP   TCP 4318  0.0.0.0/0
```

Port 22 was used for SSH / EC2 Instance Connect.

Port 4318 was opened because OTLP/HTTP uses this port. The Lambda function sends telemetry to:

http://X.X.X.X:4318

## Grafana Alloy

Grafana Alloy was installed on the EC2 instance.

Alloy works as a collector:

```
receives OTLP data from Lambda
        ↓
processes telemetry
        ↓
exports it to Grafana Cloud
```

Alloy was configured with an OTLP receiver on port 4318.

The service status was checked with:

```
sudo systemctl status alloy
```

The expected result:

```
active (running)
```

The listening port was checked with:

```
sudo ss -tulpen | grep 4318
```

The expected result is that Alloy listens on port 4318.

## Grafana Cloud

Grafana Cloud was used as the final storage and visualization platform.

In Grafana Cloud, OpenTelemetry connection details were generated. These details were used by Grafana Alloy to forward telemetry to Grafana Cloud.

Grafana Cloud provided:

```
OTLP endpoint;
instance ID / username;
access token;
hosted metrics backend.
```

## Dashboard

A Grafana dashboard was created using the metrics received from AWS Lambda.

The dashboard includes panels for:

```
Lambda invocations;
Lambda errors;
Lambda execution duration;
average duration;
error rate.
```

## Example PromQL queries:

```
lambda_invocations_total
lambda_errors_total
lambda_duration_ms_count
rate(lambda_duration_ms_sum[5m]) / rate(lambda_duration_ms_count[5m])
increase(lambda_invocations_total[5m])

```

## Why Prometheus scrape was not used

Prometheus usually collects metrics by scraping a long-running /metrics endpoint.

AWS Lambda is not a long-running server. It starts, processes an event and stops. Because of this, it is not convenient to use the classic Prometheus scrape model with Lambda.

Instead, this project uses a push-based OpenTelemetry approach:

Lambda pushes metrics → Alloy receives them → Grafana Cloud stores them
Why Grafana Alloy was used

Grafana Alloy was used as an intermediate collector between Lambda and Grafana Cloud.

This gives a cleaner architecture because Lambda does not send data directly to Grafana Cloud. Instead, Lambda sends telemetry to Alloy, and Alloy forwards it to Grafana Cloud.

Alloy can also be used later for filtering, routing, transforming and enriching telemetry data.

## Notes

This project uses an EC2 instance as a collector. For a production environment, this EC2 instance would need high availability: multiple collectors, load balancing and health checks.

For a simple educational project, one Free Tier EC2 instance is enough to demonstrate the telemetry pipeline.
