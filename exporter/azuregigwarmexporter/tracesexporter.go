// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build cgo

package azuregigwarmexporter

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	cgogeneva "go.opentelemetry.io/collector/exporter/azuregigwarmexporter/internal/cgo"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
)

// tracesExporter implements the traces exporter for Azure Geneva Warm (GigWarm) via Rust FFI.
type tracesExporter struct {
	params exporter.Settings
	cfg    *Config
	client *cgogeneva.GenevaClient
	logger *zap.Logger
}

var _ consumer.Traces = (*tracesExporter)(nil)
var _ component.Component = (*tracesExporter)(nil)

// newTracesExporter creates a new GigWarm traces exporter.
func newTracesExporter(_ context.Context, set exporter.Settings, cfg *Config) (*tracesExporter, error) {
	// Validate early to fail fast
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid azuregigwarm config: %w", err)
	}

	// Build CGO config
	cgoCfg := cgogeneva.GenevaConfig{
		Endpoint:           cfg.Endpoint,
		Environment:        cfg.Environment,
		Account:            cfg.Account,
		Namespace:          cfg.Namespace,
		Region:             cfg.Region,
		ConfigMajorVersion: cfg.ConfigMajorVersion,
		AuthMethod:         int32(cfg.AuthMethod), // 0 = MSI, 1 = Certificate
		Tenant:             cfg.Tenant,
		RoleName:           cfg.RoleName,
		RoleInstance:       cfg.RoleInstance,
	}

	// Add certificate options if needed
	if cfg.AuthMethod == Certificate {
		cgoCfg.CertPath = cfg.CertPath
		cgoCfg.CertPassword = cfg.CertPassword
	}

	client, err := cgogeneva.NewGenevaClient(cgoCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Geneva FFI client: %w", err)
	}

	return &tracesExporter{
		params: set,
		cfg:    cfg,
		client: client,
		logger: set.Logger,
	}, nil
}

// start is called by the Collector when the exporter is starting.
func (e *tracesExporter) start(_ context.Context, _ component.Host) error {
	e.logger.Info("Starting AzureGigWarm traces exporter",
		zap.String("endpoint", e.cfg.Endpoint),
		zap.String("environment", e.cfg.Environment),
		zap.String("account", e.cfg.Account),
		zap.String("namespace", e.cfg.Namespace),
		zap.String("region", e.cfg.Region),
	)
	return nil
}

// shutdown is called by the Collector when the exporter is shutting down.
func (e *tracesExporter) shutdown(_ context.Context) error {
	e.logger.Info("Shutting down AzureGigWarm traces exporter")
	if e.client != nil {
		e.client.Close()
	}
	return nil
}

// consumeTraces implements consumer.ConsumeTracesFunc signature and sends traces via Rust FFI.
func (e *tracesExporter) consumeTraces(_ context.Context, td ptrace.Traces) error {
	// Marshal to OTLP ExportTraceServiceRequest protobuf bytes
	req := ptraceotlp.NewExportRequestFromTraces(td)
	data, err := req.MarshalProto()
	if err != nil {
		return fmt.Errorf("failed to marshal traces to protobuf: %w", err)
	}

	// Encode once, then upload each batch synchronously via FFI.
	batches, err := e.client.EncodeAndCompressSpans(data)
	if err != nil {
		e.logger.Error("Failed to encode spans for Geneva Warm", zap.Error(err))
		return fmt.Errorf("failed to encode spans for Geneva Warm: %w", err)
	}
	defer batches.Close()

	n := batches.Len()
	for i := 0; i < n; i++ {
		if err := e.client.UploadBatch(batches, i); err != nil {
			e.logger.Error("Failed to upload batch to Geneva Warm", zap.Int("batch_index", i), zap.Error(err))
			return fmt.Errorf("failed to upload spans batch to Geneva Warm: %w", err)
		}
	}

	e.logger.Debug("Successfully uploaded spans to Geneva Warm",
		zap.Int("span_count", td.SpanCount()),
		zap.Int("batches", n),
	)
	return nil
}

// Capabilities implements consumer.Traces.
func (e *tracesExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// ConsumeTraces implements consumer.Traces.
func (e *tracesExporter) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	return e.consumeTraces(ctx, td)
}

// Start implements component.Component.
func (e *tracesExporter) Start(ctx context.Context, host component.Host) error {
	return e.start(ctx, host)
}

// Shutdown implements component.Component.
func (e *tracesExporter) Shutdown(ctx context.Context) error {
	return e.shutdown(ctx)
}