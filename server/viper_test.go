package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleInitialize() {
	_, _, webPA, err := Initialize("example", nil, nil, viper.New())
	if err != nil {
		panic(err)
	}

	fmt.Println(webPA.Primary.Name)
	fmt.Println(webPA.Primary.Address)
	fmt.Println(webPA.Primary.LogConnectionState)

	fmt.Println(webPA.Alternate.Name)
	fmt.Println(webPA.Alternate.Address)
	fmt.Println(webPA.Alternate.LogConnectionState)

	fmt.Println(webPA.Health.Name)
	fmt.Println(webPA.Health.Address)
	fmt.Println(webPA.Health.LogInterval)
	fmt.Println(webPA.Health.Options)

	// Output:
	// example
	// localhost:10010
	// true
	// example.alternate
	// :10011
	// false
	// example.health
	// :9001
	// 45s
	// [TotalRequests TotalResponses SomeOtherStat]
}

func ExampleInitializeWithFlags() {
	var (
		f = pflag.NewFlagSet("applicationName", pflag.ContinueOnError)
		v = viper.New()

		// simulates passing `-f example` on the command line
		_, _, webPA, err = Initialize("applicationName", []string{"-f", "example"}, f, v)
	)

	if err != nil {
		panic(err)
	}

	fmt.Println(webPA.Primary.Name)
	fmt.Println(webPA.Primary.Address)
	fmt.Println(webPA.Primary.LogConnectionState)

	fmt.Println(webPA.Alternate.Name)
	fmt.Println(webPA.Alternate.Address)
	fmt.Println(webPA.Alternate.LogConnectionState)

	fmt.Println(webPA.Health.Name)
	fmt.Println(webPA.Health.Address)
	fmt.Println(webPA.Health.LogInterval)
	fmt.Println(webPA.Health.Options)

	// Output:
	// applicationName
	// localhost:10010
	// true
	// applicationName.alternate
	// :10011
	// false
	// applicationName.health
	// :9001
	// 45s
	// [TotalRequests TotalResponses SomeOtherStat]
}

func TestConfigureWhenParseError(t *testing.T) {
	var (
		assert = assert.New(t)

		f   = pflag.NewFlagSet("applicationName", pflag.ContinueOnError)
		v   = viper.New()
		err = Configure("applicationName", []string{"-unknown"}, f, v)
	)

	assert.NotNil(err)
}

func TestInitializeWhenConfigureError(t *testing.T) {
	var (
		assert = assert.New(t)

		f = pflag.NewFlagSet("applicationName", pflag.ContinueOnError)
		v = viper.New()

		logger, registry, webPA, err = Initialize("applicationName", []string{"-unknown"}, f, v)
	)

	assert.NotNil(logger)
	assert.Nil(registry)
	assert.Nil(webPA)
	assert.NotNil(err)
}

func TestInitializeWhenReadInConfigError(t *testing.T) {
	var (
		assert = assert.New(t)

		f = pflag.NewFlagSet("applicationName", pflag.ContinueOnError)
		v = viper.New()

		logger, registry, webPA, err = Initialize("applicationName", []string{"-f", "thisfiledoesnotexist"}, f, v)
	)

	assert.NotNil(logger)
	assert.Nil(registry)
	assert.Nil(webPA)
	assert.NotNil(err)
}

func TestInitializeWhenWebPAUnmarshalError(t *testing.T) {
	var (
		assert = assert.New(t)

		f = pflag.NewFlagSet("invalidWebPA", pflag.ContinueOnError)
		v = viper.New()

		logger, registry, webPA, err = Initialize("invalidWebPA", nil, f, v)
	)

	assert.NotNil(logger)
	assert.Nil(registry)
	assert.Nil(webPA)
	assert.NotNil(err)
}

func TestInitializeWhenWebPANewLoggerError(t *testing.T) {
	var (
		assert = assert.New(t)

		f = pflag.NewFlagSet("invalidLog", pflag.ContinueOnError)
		v = viper.New()

		logger, registry, webPA, err = Initialize("invalidLog", nil, f, v)
	)

	assert.NotNil(logger)
	assert.NotNil(registry)
	assert.NotNil(webPA)
	assert.Nil(err)
}

func TestInitializeMetrics(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		v       = viper.New()
		w       = new(WebPA)
	)

	v.SetConfigType("json")
	require.NoError(v.ReadConfig(strings.NewReader(`
		{
			"metric": {
				"address": ":8080",
				"metricsOptions": {
					"namespace": "foo",
					"subsystem": "bar"
				}
			}
		}
	`)))

	require.NoError(v.Unmarshal(w))

	assert.Equal("foo", w.Metric.MetricsOptions.Namespace)
	assert.Equal("bar", w.Metric.MetricsOptions.Subsystem)
}
