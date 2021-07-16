package consul

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/mitchellh/mapstructure"
	"github.com/xmidt-org/argus/chrysom"
	"github.com/xmidt-org/argus/model"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/webpa-common/service"
)

// datacenterWatcher checks if datacenters have been updated, based on an interval.
type datacenterWatcher struct {
	logger              log.Logger
	environment         Environment
	options             Options
	inactiveDatacenters map[string]bool
	chrysomClient       *chrysom.Client
	consulWatchInterval time.Duration
	lock                sync.RWMutex
}

type datacenterFilter struct {
	Name     string
	Inactive bool
}

var (
	defaultLogger        = log.NewNopLogger()
	defaultWatchInterval = 5 * time.Minute
)

func newDatacenterWatcher(logger log.Logger, environment Environment, options Options) (*datacenterWatcher, error) {

	if logger == nil {
		logger = defaultLogger
	}

	if options.DatacenterWatchInterval <= 0 {
		//default consul interval is 5m
		options.DatacenterWatchInterval = defaultWatchInterval
	}

	datacenterWatcher := &datacenterWatcher{
		consulWatchInterval: options.DatacenterWatchInterval,
		logger:              logger,
		options:             options,
		environment:         environment,
		inactiveDatacenters: make(map[string]bool),
	}

	if len(options.Chrysom.Bucket) > 0 {
		if options.Chrysom.Listen.PullInterval <= 0 {
			return nil, errors.New("chrysom pull interval cannot be 0")
		}

		// only chrysom client uses the provider for metrics
		if environment.Provider() == nil {
			return nil, errors.New("must pass in a metrics provider")
		}

		var datacenterListenerFunc chrysom.ListenerFunc = func(items chrysom.Items) {
			updateInactiveDatacenters(items, datacenterWatcher.inactiveDatacenters, &datacenterWatcher.lock, logger)
		}

		options.Chrysom.Listen.Listener = datacenterListenerFunc
		options.Chrysom.Logger = logger

		m := &chrysom.Measures{
			Polls: environment.Provider().NewCounterVec(chrysom.PollCounter),
		}
		chrysomClient, err := chrysom.NewClient(options.Chrysom, m, getLogger, logging.WithLogger)

		if err != nil {
			return nil, err
		}

		//create chrysom client and start it
		datacenterWatcher.chrysomClient = chrysomClient
		datacenterWatcher.chrysomClient.Start(context.Background())
		logger.Log(level.Key(), level.DebugValue(), logging.MessageKey(), "started chrysom, argus client")
	}

	//start consul watch
	ticker := time.NewTicker(datacenterWatcher.consulWatchInterval)
	go datacenterWatcher.watchDatacenters(ticker)
	logger.Log(level.Key(), level.DebugValue(), logging.MessageKey(), "started consul datacenter watch")

	return datacenterWatcher, nil

}

func (d *datacenterWatcher) stop() {
	if d.chrysomClient != nil {
		d.chrysomClient.Stop(context.Background())
	}
}

func (d *datacenterWatcher) watchDatacenters(ticker *time.Ticker) {
	for {
		select {
		case <-d.environment.Closed():
			ticker.Stop()
			d.stop()
			return
		case <-ticker.C:
			datacenters, err := getDatacenters(d.logger, d.environment.Client(), d.options)

			if err != nil {
				// getDatacenters function logs the error, but a metric should be added
				continue
			}

			d.updateInstancers(datacenters)

		}

	}
}

func (d *datacenterWatcher) updateInstancers(datacenters []string) {
	keys := make(map[string]bool)
	instancersToAdd := make(service.Instancers)

	currentInstancers := d.environment.Instancers()

	for _, w := range d.options.watches() {
		if w.CrossDatacenter {
			for _, datacenter := range datacenters {

				createNewInstancer(keys, instancersToAdd, currentInstancers, d, datacenter, w)
			}
		}
	}

	d.environment.UpdateInstancers(keys, instancersToAdd)

}

func updateInactiveDatacenters(items []model.Item, inactiveDatacenters map[string]bool, lock *sync.RWMutex, logger log.Logger) {
	chrysomMap := make(map[string]bool)
	for _, item := range items {

		var df datacenterFilter

		// decode database item data into datacenter filter structure
		err := mapstructure.Decode(item.Data, &df)

		if err != nil {
			logger.Log(level.Key(), level.ErrorValue(), logging.MessageKey(), "failed to decode database results into datacenter filter struct")
			continue
		}

		if df.Inactive {
			chrysomMap[df.Name] = true
			lock.Lock()
			inactiveDatacenters[df.Name] = true
			lock.Unlock()
		}
	}

	lock.Lock()

	for key := range inactiveDatacenters {
		_, ok := chrysomMap[key]

		if !ok {
			delete(inactiveDatacenters, key)
		}
	}

	lock.Unlock()
}

func createNewInstancer(keys map[string]bool, instancersToAdd service.Instancers, currentInstancers service.Instancers, dw *datacenterWatcher, datacenter string, w Watch) {
	//check if datacenter is part of inactive datacenters list
	dw.lock.RLock()
	_, found := dw.inactiveDatacenters[datacenter]
	dw.lock.RUnlock()

	if found {
		dw.logger.Log(level.Key(), level.InfoValue(), logging.MessageKey(), "datacenter set as inactive", "datacenter name: ", datacenter)
		return
	}

	w.QueryOptions.Datacenter = datacenter

	// create keys for all datacenters + watched services
	key := newInstancerKey(w)
	keys[key] = true

	// don't create new instancer if it is already saved in environment's instancers
	if currentInstancers.Has(key) {
		return
	}

	// don't create new instancer if it was already created and added to the new instancers map
	if instancersToAdd.Has(key) {
		dw.logger.Log(level.Key(), level.WarnValue(), logging.MessageKey(), "skipping duplicate watch", "service", w.Service, "tags", w.Tags, "passingOnly", w.PassingOnly, "datacenter", w.QueryOptions.Datacenter)
		return
	}

	// create new instancer and add it to the map of instancers to add
	instancersToAdd.Set(key, newInstancer(dw.logger, dw.environment.Client(), w))
}

func getLogger(ctx context.Context) log.Logger {
	logger := log.With(logging.GetLogger(ctx), "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	return logger
}
