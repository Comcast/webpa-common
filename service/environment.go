package service

import (
	"errors"
	"io"
	"sync"

	"github.com/go-kit/kit/sd"
	"github.com/xmidt-org/webpa-common/xmetrics"
)

var (
	// ErrIncomplete is returned by platform-specific code that creates environments
	// whenever the configuration is incomplete: it contains connection information
	// but no watches or registrations.
	ErrIncomplete = errors.New("No watches or registrations configured")
)

// NopCloser is a closer function that does nothing.  It always returns a nil error.  Useful
// for testing.  Also used internally for the Environment's closer function in place of nil.
func NopCloser() error { return nil }

// Environment represents everything known about a service discovery backend.  It also
// provide a central handle for tasks related to service discovery, such as Accessor hashing.
type Environment interface {
	sd.Registrar
	io.Closer

	// IsRegistered tests if the given instance is registered in this environment.  Useful for
	// determining if an arbitrary instance refers to this process.
	IsRegistered(string) bool

	// DefaultScheme is the default URI scheme to assume for discovered service instances.  This is
	// typically driven by configuration.
	DefaultScheme() string

	// Instancers returns a copy of the internal set of Instancers this environment is configured to watch.
	// Changing the returned Instancers will not result in changing this Environment's state.
	Instancers() Instancers

	// UpdateInstancers configures the set of sd.Instancer objects for use in the environment.
	UpdateInstancers(currentKeys map[string]bool, instancersToAdd Instancers)

	// AccessorFactory returns the creation strategy for Accessors used in this environment.
	// Typically, this factory is set via configuration by some external source.
	AccessorFactory() AccessorFactory

	// Closed returns a channel that is closed when this Environment in closed.
	Closed() <-chan struct{}

	// Provider returns the metrics provider that is associated with this environment
	// Mainly used for the argus chrysom client
	Provider() xmetrics.Registry
}

// Option represents a service discovery option for configuring an Environment
type Option func(*environment)

// WithDefaultScheme configures the default URI scheme for discovered instances that do not
// specify a scheme.  Some service discovery backends do not have a way to advertise a particular
// scheme that is revealed as part of the discovered instances.
func WithDefaultScheme(s string) Option {
	return func(e *environment) {
		if len(s) > 0 {
			e.defaultScheme = s
		} else {
			e.defaultScheme = DefaultScheme
		}
	}
}

// WithRegistrars configures the mapping of sd.Registrar objects to use for service
// advertisement.
func WithRegistrars(r Registrars) Option {
	return func(e *environment) {
		e.registrars = r
	}
}

// WithInstancers configures the set of sd.Instancer objects for use in the environment.
// The Instancers may be nil or empty for applications which have no need of monitoring
// discovered services.
func WithInstancers(i Instancers) Option {
	return func(e *environment) {
		e.instancers = i.Copy()
	}
}

// WithAccessorFactory configures the creation strategy for Accessor objects.  By default,
// DefaultAccessorFactory is used.  Passing nil via this option sets (or resets) the environment
// back to using the DefaultAccessorFactory.
func WithAccessorFactory(af AccessorFactory) Option {
	return func(e *environment) {
		if af == nil {
			e.accessorFactory = DefaultAccessorFactory
		} else {
			e.accessorFactory = af
		}
	}
}

// WithCloser configures the function used to completely shut down the service discover backend.
// By default, NopCloser is used.  Passing a nil function for this option sets (or resets)
// the closer back to the NopCloser.
func WithCloser(f func() error) Option {
	return func(e *environment) {
		if f == nil {
			e.closer = NopCloser
		} else {
			e.closer = f
		}
	}
}

// WithProvider configures the metrics provider for the environment
func WithProvider(p xmetrics.Registry) Option {
	return func(e *environment) {
		if p != nil {
			e.provider = p
		}
	}
}

// NewEnvironment constructs a new service discovery client environment.  It is possible to construct
// an environment without any Registrars or Instancers, which essentially makes a no-op environment.
func NewEnvironment(options ...Option) Environment {
	e := &environment{
		defaultScheme:   DefaultScheme,
		accessorFactory: DefaultAccessorFactory,
		closer:          NopCloser,
		closed:          make(chan struct{}),
	}

	for _, o := range options {
		o(e)
	}

	return e
}

type environment struct {
	defaultScheme   string
	registrars      Registrars
	instancers      Instancers
	accessorFactory AccessorFactory
	provider        xmetrics.Registry

	lock      sync.RWMutex
	closeOnce sync.Once
	closer    func() error
	closed    chan struct{}
}

func (e *environment) IsRegistered(instance string) bool {
	return e.registrars.Has(instance)
}

func (e *environment) DefaultScheme() string {
	return e.defaultScheme
}

func (e *environment) Instancers() Instancers {
	e.lock.RLock()
	instancersCopy := e.instancers.Copy()
	e.lock.RUnlock()
	return instancersCopy
}

func (e *environment) UpdateInstancers(currentKeys map[string]bool, instancersToAdd Instancers) {
	// add new instancers

	for key, value := range instancersToAdd {
		e.lock.Lock()
		e.instancers.Set(key, value)
		e.lock.Unlock()
	}

	// remove outdated instancers
	e.lock.Lock()
	for key := range e.instancers {
		if _, ok := currentKeys[key]; !ok {
			i, found := e.instancers.Get(key)

			if found {
				i.Stop()
				delete(e.instancers, key)
			}
		}
	}
	e.lock.Unlock()
}

func (e *environment) AccessorFactory() AccessorFactory {
	return e.accessorFactory
}

func (e *environment) Register() {
	e.registrars.Register()
}

func (e *environment) Deregister() {
	e.registrars.Deregister()
}

func (e *environment) Closed() <-chan struct{} {
	return e.closed
}

// Close completely shuts down this environment.  Any registrars are deregistered, all instancers are stopped,
// the Closed() channel is closed, and any configured closer function is invoked.  This method is idempotent
// and safe for concurrent execution.
func (e *environment) Close() (err error) {
	e.closeOnce.Do(func() {
		e.Deregister()
		for _, v := range e.instancers {
			v.Stop()
		}

		close(e.closed)
		err = e.closer()
	})

	return
}

func (e *environment) Provider() xmetrics.Registry {
	return e.provider
}
