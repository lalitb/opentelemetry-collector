// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build cgo

package azuregigwarmexporter

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
)

var (
	Type = component.MustNewType("azuregigwarm")
)

const (
	stability = component.StabilityLevelAlpha
)

var (
	errUnexpectedConfigurationType = errors.New("failed to cast configuration to AzureGigWarm Config")
)

type factory struct{}

// NewFactory creates an exporter factory for Azure Geneva Warm.
func NewFactory() exporter.Factory {
	f := &factory{}
	return exporter.NewFactory(
		Type,
		f.createDefaultConfig,
		exporter.WithLogs(f.createLogsExporter, stability),
	)
}

// createDefaultConfig creates the default exporter configuration.
func (f *factory) createDefaultConfig() component.Config {
	return &Config{}
}

// createLogsExporter creates a logs exporter based on the config.
func (f *factory) createLogsExporter(ctx context.Context, set exporter.Settings, c component.Config) (exporter.Logs, error) {
	cfg, ok := c.(*Config)
	if !ok {
		return nil, errUnexpectedConfigurationType
	}

	exp, err := newLogsExporter(ctx, set, cfg)
	if err != nil {
		return nil, err
	}

	// Return the exporter directly without using exporterhelper
	return exp, nil
}
