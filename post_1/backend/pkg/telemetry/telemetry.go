package telemetry

import (
	"context"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Telemetry struct {
	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter

	// HTTP metrics
	RequestsTotal    metric.Int64Counter
	RequestDuration  metric.Float64Histogram
	RequestsInFlight metric.Int64UpDownCounter
	ResponseSize     metric.Int64Histogram
}

func NewTelemetry(logger *slog.Logger) (*Telemetry, error) {
	// Create Prometheus exporter
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExporter),
	)

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter := meterProvider.Meter("personal-website")

	// Create metrics
	requestsTotal, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	requestsInFlight, err := meter.Int64UpDownCounter(
		"http_requests_in_flight",
		metric.WithDescription("Current number of HTTP requests being processed"),
	)
	if err != nil {
		return nil, err
	}

	responseSize, err := meter.Int64Histogram(
		"http_response_size_bytes",
		metric.WithDescription("HTTP response size in bytes"),
	)
	if err != nil {
		return nil, err
	}

	return &Telemetry{
		meterProvider:    meterProvider,
		meter:            meter,
		RequestsTotal:    requestsTotal,
		RequestDuration:  requestDuration,
		RequestsInFlight: requestsInFlight,
		ResponseSize:     responseSize,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	return t.meterProvider.Shutdown(ctx)
}

// ResponseWriter wraps http.ResponseWriter to capture metrics data
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Written    int64
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.Written += int64(n)
	return n, err
}

// NewResponseWriter creates a new ResponseWriter with default status code
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}