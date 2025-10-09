// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build cgo

package azuregigwarmexporter

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	cgogeneva "go.opentelemetry.io/collector/exporter/azuregigwarmexporter/internal/cgo"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.uber.org/zap"
)

// logsExporter implements the logs exporter for Azure Geneva Warm (GigWarm) via Rust FFI.
type logsExporter struct {
	params exporter.Settings
	cfg    *Config
	client *cgogeneva.GenevaClient
	logger *zap.Logger
}

// logsExporter no longer needs to implement consumer.Logs or component.Component
// because exporterhelper handles those interfaces

// newLogsExporter creates a new GigWarm logs exporter.
func newLogsExporter(_ context.Context, set exporter.Settings, cfg *Config) (*logsExporter, error) {
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

	return &logsExporter{
		params: set,
		cfg:    cfg,
		client: client,
		logger: set.Logger,
	}, nil
}

// start is called by the Collector when the exporter is starting.
func (e *logsExporter) start(_ context.Context, _ component.Host) error {
	e.logger.Info("Starting AzureGigWarm exporter",
		zap.String("endpoint", e.cfg.Endpoint),
		zap.String("environment", e.cfg.Environment),
		zap.String("account", e.cfg.Account),
		zap.String("namespace", e.cfg.Namespace),
		zap.String("region", e.cfg.Region),
	)
	return nil
}

// shutdown is called by the Collector when the exporter is shutting down.
func (e *logsExporter) shutdown(_ context.Context) error {
	e.logger.Info("Shutting down AzureGigWarm exporter")
	if e.client != nil {
		e.client.Close()
	}
	return nil
}

// pushLogs implements the push function for exporterhelper and sends logs via Rust FFI.
func (e *logsExporter) pushLogs(_ context.Context, ld plog.Logs) error {
	// Marshal to OTLP ExportLogsServiceRequest protobuf bytes
	req := plogotlp.NewExportRequestFromLogs(ld)
	data, err := req.MarshalProto()
	if err != nil {
		return fmt.Errorf("failed to marshal logs to protobuf: %w", err)
	}

	// Encode once, then upload each batch synchronously via FFI.
	batches, err := e.client.EncodeAndCompressLogs(data)
	if err != nil {
		e.logger.Error("Failed to encode logs for Geneva Warm", zap.Error(err))
		return fmt.Errorf("failed to encode logs for Geneva Warm: %w", err)
	}
	defer batches.Close()

	n := batches.Len()

	// Upload batches concurrently for better throughput
	errChan := make(chan error, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if err := e.client.UploadBatch(batches, index); err != nil {
				e.logger.Error("Failed to upload batch to Geneva Warm", zap.Int("batch_index", index), zap.Error(err))
				errChan <- fmt.Errorf("failed to upload logs batch %d to Geneva Warm: %w", index, err)
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	if err := <-errChan; err != nil {
		return err
	}

	e.logger.Debug("Successfully uploaded logs to Geneva Warm",
		zap.Int("log_records", ld.LogRecordCount()),
		zap.Int("batches", n),
	)
	return nil
}

// These interface methods are no longer needed because exporterhelper wraps the exporter
// and handles the consumer.Logs and component.Component interfaces
