package fanout

import (
	"time"

	"github.com/Comcast/webpa-common/xhttp"
	"github.com/Comcast/webpa-common/xhttp/xcontext"
	"github.com/Comcast/webpa-common/xhttp/xtimeout"
	gokithttp "github.com/go-kit/kit/transport/http"
	"github.com/justinas/alice"
)

const (
	DefaultFanoutTimeout time.Duration = 45 * time.Second
	DefaultClientTimeout time.Duration = 30 * time.Second
	DefaultConcurrency                 = 1000
)

// Configuration defines the configuration structure for externally configuring a fanout.
type Configuration struct {
	// Endpoints are the URLs for each endpoint to fan out to.  If unset, the default is supplied
	// by application code, which is normally a set of endpoints driven by service discovery.
	Endpoints []string `json:"endpoints,omitempty"`

	// Authorization is the Basic Auth token.  There is no default for this field.
	Authorization string `json:"authorization"`

	// FanoutTimeout is the timeout for the entire fanout operation.  If not supplied, DefaultFanoutTimeout is used.
	FanoutTimeout time.Duration `json:"fanoutTimeout"`

	// Concurrency is the maximum number of concurrent fanouts allowed.  If this is not set, DefaultConcurrency is used.
	Concurrency int `json:"concurrency"`
}

func (c *Configuration) endpoints() []string {
	if c != nil {
		return c.Endpoints
	}

	return nil
}

func (c *Configuration) authorization() string {
	if c != nil && len(c.Authorization) > 0 {
		return c.Authorization
	}

	return ""
}

func (c *Configuration) fanoutTimeout() time.Duration {
	if c != nil && c.FanoutTimeout > 0 {
		return c.FanoutTimeout
	}

	return DefaultFanoutTimeout
}

func (c *Configuration) concurrency() int {
	if c != nil && c.Concurrency > 0 {
		return c.Concurrency
	}

	return DefaultConcurrency
}

// NewChain constructs an Alice constructor Chain from a set of fanout options and zero or
// more application-layer request functions.
func NewChain(c Configuration, rf ...gokithttp.RequestFunc) alice.Chain {
	return alice.New(
		xtimeout.NewConstructor(xtimeout.Options{
			Timeout: c.fanoutTimeout(),
		}),
		xcontext.Populate(rf...),
		xhttp.Busy(c.concurrency()),
	)
}
