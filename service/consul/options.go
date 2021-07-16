package consul

import (
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xmidt-org/argus/chrysom"
)

const DefaultDatacenterRetries = 10

type Watch struct {
	Service         string           `json:"service,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	PassingOnly     bool             `json:"passingOnly"`
	CrossDatacenter bool             `json:"crossDatacenter"`
	QueryOptions    api.QueryOptions `json:"queryOptions"`
}

type Options struct {
	Client                  *api.Config                    `json:"client"`
	Chrysom                 chrysom.ClientConfig           `json:"chrysom"`
	DisableGenerateID       bool                           `json:"disableGenerateID"`
	DatacenterRetries       int                            `json:"datacenterRetries"`
	DatacenterWatchInterval time.Duration                  `json:"datacenterWatchInterval"`
	Registrations           []api.AgentServiceRegistration `json:"registrations,omitempty"`
	Watches                 []Watch                        `json:"watches,omitempty"`
}

func (o *Options) config() *api.Config {
	if o != nil && o.Client != nil {
		return o.Client
	}

	return api.DefaultConfig()
}

func (o *Options) disableGenerateID() bool {
	if o != nil {
		return o.DisableGenerateID
	}

	return false
}

func (o *Options) datacenterRetries() int {
	if o != nil && o.DatacenterRetries > 0 {
		return o.DatacenterRetries
	}

	return DefaultDatacenterRetries
}

func (o *Options) registrations() []api.AgentServiceRegistration {
	if o != nil && len(o.Registrations) > 0 {
		return o.Registrations
	}

	return nil
}

func (o *Options) watches() []Watch {
	if o != nil && len(o.Watches) > 0 {
		return o.Watches
	}

	return nil
}
