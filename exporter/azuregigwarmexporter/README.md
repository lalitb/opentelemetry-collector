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

Static linking: the Rust FFI library is statically linked into the collector binary by default, so no DYLD_LIBRARY_PATH / LD_LIBRARY_PATH adjustments are required. (Dynamic linking is only needed if you deliberately change `geneva_ffi.go` to reference the `.dylib` explicitly.)

## Send a Test Log

Using [otel-cli](https://github.com/equinix-labs/otel-cli):

```bash
otel-cli --endpoint localhost:4317 --otlp-insecure logs --body "hello gigwarm" --severity INFO
```

## Send Test Spans

Using [otel-cli](https://github.com/equinix-labs/otel-cli):

```bash
otel-cli --endpoint localhost:4317 --otlp-insecure span --name "test-span" --service "my-service"
```

To send a span with additional attributes:

```bash
otel-cli --endpoint localhost:4317 --otlp-insecure span \
  --name "test-span" \
  --service "my-service" \
  --attrs "http.method=GET,http.url=/api/test"
```

Note: To enable span export, update the collector config to include a `traces` pipeline:

```yaml
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [azuregigwarm]
```

If configuration values are missing the exporter will emit validation errors on startup.
