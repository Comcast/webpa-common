package health

import (
	"encoding/json"
	"fmt"
	"github.com/Comcast/webpa-common/logging"
	"net/http"
	"sync"
	"time"
)

// StatsListener receives Stats on regular intervals.
type StatsListener interface {
	// OnStats is called with a copy of the health's stats map
	// at regular intervals.
	OnStats(Stats)
}

// StatsListenerFunc is a function type that implements StatsListener.
type StatsListenerFunc func(Stats)

func (f StatsListenerFunc) OnStats(stats Stats) {
	f(stats)
}

// Monitor is the basic interface implemented by health event sinks
type Monitor interface {
	SendEvent(HealthFunc)

	// HACK HACK HACK
	// This should be moved to another package
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Health is the central type of this package.  It defines and endpoint for tracking
// and updating various statistics.  It also dispatches events to one or more StatsListeners
// at regular intervals.
type Health struct {
	stats            Stats
	statDumpInterval time.Duration
	log              logging.Logger
	events           chan HealthFunc
	statsListeners   []StatsListener
	memInfoReader    *MemInfoReader
	once             sync.Once
}

var _ Monitor = (*Health)(nil)

// AddStatsListener adds a new listener to this Health.  This method
// is asynchronous.  The listener will eventually receive events, but callers
// should not assume events will be dispatched immediately after this method call.
func (h *Health) AddStatsListener(listener StatsListener) {
	h.SendEvent(func(stat Stats) {
		h.statsListeners = append(h.statsListeners, listener)
	})
}

// SendEvent dispatches a HealthFunc to the internal event queue
func (h *Health) SendEvent(healthFunc HealthFunc) {
	h.events <- healthFunc
}

// New creates a Health object with the given statistics.
func New(interval time.Duration, log logging.Logger, options ...Option) *Health {
	initialStats := commonStats.Clone()
	initialStats.Apply(options...)

	return &Health{
		stats:            initialStats,
		statDumpInterval: interval,
		log:              log,
		memInfoReader:    &MemInfoReader{},
	}
}

// Run executes this Health object.  This method is idempotent:  once a
// Health object is Run, it cannot be Run again.
func (h *Health) Run(waitGroup *sync.WaitGroup, shutdown <-chan struct{}) error {
	h.once.Do(func() {
		h.log.Debug("Health Monitor Started")
		h.events = make(chan HealthFunc, 100)

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			ticker := time.NewTicker(h.statDumpInterval)
			defer ticker.Stop()
			defer h.log.Debug("Health Monitor Stopped")
			defer close(h.events)

			for {
				select {
				case <-shutdown:
					return

				case hf := <-h.events:
					hf(h.stats)

				case <-ticker.C:
					h.stats.UpdateMemory(h.memInfoReader)
					dispatchStats := h.stats.Clone()
					for _, statsListener := range h.statsListeners {
						statsListener.OnStats(dispatchStats)
					}
				}
			}
		}()
	})

	return nil
}

func (h *Health) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	output := make(chan Stats, 1)
	defer close(output)

	h.SendEvent(func(stats Stats) {
		stats.UpdateMemory(h.memInfoReader)
		output <- stats.Clone()
	})

	stats := <-output
	response.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(stats)

	// TODO: leverage the standard error writing elsewhere in webpa-common
	if err != nil {
		h.log.Error("Could not marshal stats: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(response, `{"message": "%s"}\n`, err.Error())
	} else {
		fmt.Fprintf(response, "%s", data)
	}
}
