# Azure GigWarm Exporter (Experimental)

Minimal instructions to build and test the Azure GigWarm (Geneva Warm) logs exporter that uses a Rust FFI bridge.

## Build (enable exporter)

```bash
make AZUREGIGWARM=1 otelcorecol
# Result: ./bin/otelcorecol_$(go env GOOS)_$(go env GOARCH)
```

(Without `AZUREGIGWARM=1` the binary builds without the real exporter and a stub reports a CGO requirement.)

## Example Collector Config

Save as: `examples/local/otel-gigwarm.yaml`

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:

exporters:
  azuregigwarm:
    endpoint: https://YOUR_GENEVA_ENDPOINT
    environment: prod
    account: your_account
    namespace: your_namespace
    region: westus2
    config_major_version: 1
    auth_method: 0          # 0=msi, 1=certificate
    tenant: your_tenant
    role_name: your_role
    role_instance: instance-1
    # For certificate auth:
    # auth_method: 1
    # cert_path: /path/to/cert.pfx
    # cert_password: secret

processors:
  batch:

service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [azuregigwarm]
```

## Run

```bash
./bin/otelcorecol_$(go env GOOS)_$(go env GOARCH) --config examples/local/otel-gigwarm.yaml
```

(If the dynamic library cannot be found at runtime, export DYLD_LIBRARY_PATH or LD_LIBRARY_PATH pointing to `exporter/azuregigwarmexporter/geneva_ffi_bridge/target/release`.)

## Send a Test Log

Using [otel-cli](https://github.com/equinix-labs/otel-cli):

```bash
otel-cli --endpoint localhost:4317 --otlp-insecure logs --body "hello gigwarm" --severity INFO
```

If configuration values are missing the exporter will emit validation errors on startup.
