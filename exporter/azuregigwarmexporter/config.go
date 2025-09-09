// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azuregigwarmexporter // import "dev.azure.com/msazure/one/_git/strato.git/exporter/azuregigwarmexporter"

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
)

// AuthMethod represents the authentication method for Geneva Warm
type AuthMethod int

const (
	// MSI uses Managed Service Identity
	MSI AuthMethod = iota
	// Certificate uses certificate-based authentication
	Certificate
)

// String returns the string representation of AuthMethod
func (a AuthMethod) String() string {
	switch a {
	case MSI:
		return "msi"
	case Certificate:
		return "certificate"
	default:
		return "unknown"
	}
}

// Config defines configuration for the Azure Geneva Warm exporter.
//
// This exporter sends OTLP Log data to Azure Geneva (Warm path) using a Rust FFI uploader.
type Config struct {
	// Geneva-specific configuration (required)
	Endpoint           string     `mapstructure:"endpoint"`
	Environment        string     `mapstructure:"environment"`
	Account            string     `mapstructure:"account"`
	Namespace          string     `mapstructure:"namespace"`
	Region             string     `mapstructure:"region"`
	ConfigMajorVersion uint32     `mapstructure:"config_major_version"`
	AuthMethod         AuthMethod `mapstructure:"auth_method"`
	Tenant             string     `mapstructure:"tenant"`
	RoleName           string     `mapstructure:"role_name"`
	RoleInstance       string     `mapstructure:"role_instance"`

	// Certificate auth parameters (optional; required only when AuthMethod == Certificate)
	CertPath     string `mapstructure:"cert_path"`
	CertPassword string `mapstructure:"cert_password"`

	// prevent unkeyed literal initialization
	_ struct{}
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New(`requires a non-empty "endpoint"`)
	}
	if cfg.Environment == "" {
		return errors.New(`requires a non-empty "environment"`)
	}
	if cfg.Account == "" {
		return errors.New(`requires a non-empty "account"`)
	}
	if cfg.Namespace == "" {
		return errors.New(`requires a non-empty "namespace"`)
	}
	if cfg.Region == "" {
		return errors.New(`requires a non-empty "region"`)
	}
	if cfg.Tenant == "" {
		return errors.New(`requires a non-empty "tenant"`)
	}
	if cfg.RoleName == "" {
		return errors.New(`requires a non-empty "role_name"`)
	}
	if cfg.RoleInstance == "" {
		return errors.New(`requires a non-empty "role_instance"`)
	}
	if cfg.AuthMethod != MSI && cfg.AuthMethod != Certificate {
		return fmt.Errorf(`invalid auth_method: %d (must be 0 for MSI or 1 for Certificate)`, cfg.AuthMethod)
	}
	if cfg.AuthMethod == Certificate {
		if cfg.CertPath == "" {
			return errors.New(`requires a non-empty "cert_path" when auth_method == certificate`)
		}
		// cert_password can be empty if the cert is not password protected, so no hard check here.
	}
	return nil
}
