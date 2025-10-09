module go.opentelemetry.io/collector/testbed

go 1.24

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.0.0
	github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed v0.0.0
	go.opentelemetry.io/collector/component v0.0.0
	go.opentelemetry.io/collector/consumer v0.0.0
	go.opentelemetry.io/collector/pdata v0.0.0
)

// Use local collector modules
replace go.opentelemetry.io/collector => ../

replace go.opentelemetry.io/collector/component => ../component

replace go.opentelemetry.io/collector/consumer => ../consumer

replace go.opentelemetry.io/collector/pdata => ../pdata

// Point to contrib testbed (you may need to adjust this path)
replace github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed => ../../../opentelemetry-collector-contrib/testbed/testbed

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/common => ../../../opentelemetry-collector-contrib/internal/common
