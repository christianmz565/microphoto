// Package metrics provides OpenTelemetry metrics instrumentation and a Prometheus metrics server.
package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Metrics holds OpenTelemetry metric instruments for task processing.
type Metrics struct {
	TasksProcessed metric.Int64Counter
	TaskDuration   metric.Float64Histogram
	TaskTimeouts   metric.Int64Counter
}

// InitMetrics initializes OpenTelemetry meters and registers them with a Prometheus exporter.
func InitMetrics(serviceName string) (*Metrics, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("creating prometheus exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	meter := otel.Meter("microphoto")

	tasksProcessed, err := meter.Int64Counter("tasks_processed_total",
		metric.WithDescription("Total number of tasks processed"),
	)
	if err != nil {
		return nil, err
	}

	taskDuration, err := meter.Float64Histogram("task_duration_seconds",
		metric.WithDescription("Average processing time per task"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	taskTimeouts, err := meter.Int64Counter("task_timeouts_total",
		metric.WithDescription("Total number of task timeouts detected"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		TasksProcessed: tasksProcessed,
		TaskDuration:   taskDuration,
		TaskTimeouts:   taskTimeouts,
	}, nil
}

// StartMetricsServer starts an HTTP server exposing Prometheus metrics.
func StartMetricsServer(port int) {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		srv := &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}
		fmt.Printf("Metrics server listening on :%d/metrics\n", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server failed: %v\n", err)
		}
	}()
}

// RecordTaskProcessed increments the processed tasks counter.
func (m *Metrics) RecordTaskProcessed(ctx context.Context, workerID, taskType string) {
	m.TasksProcessed.Add(ctx, 1, metric.WithAttributes(
		semconv.ServiceInstanceID(workerID),
		attribute.String("task_type", taskType),
	))
}

// RecordTaskDuration records the duration of a task.
func (m *Metrics) RecordTaskDuration(ctx context.Context, workerID string, duration float64) {
	m.TaskDuration.Record(ctx, duration, metric.WithAttributes(
		semconv.ServiceInstanceID(workerID),
	))
}

// RecordTaskTimeout increments the timeout counter.
func (m *Metrics) RecordTaskTimeout(ctx context.Context, workerID string) {
	m.TaskTimeouts.Add(ctx, 1, metric.WithAttributes(
		semconv.ServiceInstanceID(workerID),
	))
}
