package profiler

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const (
	minPort = 1024
	maxPort = 49151

	byte = 1 << (10 * iota)
	kilobyte
	megabyte
	gigabyte
	terabyte
)

// Service opts holds configuration options for the profiler service.
type ServiceOpts struct {
	Port          int
	StatsInterval time.Duration
	Datadir       string
}

func (o ServiceOpts) validate() error {
	if len(o.Datadir) == 0 {
		return fmt.Errorf("missing profiler datadir")
	}
	if o.Port < minPort || o.Port > maxPort {
		return fmt.Errorf("port must be in range [%d, %d]", minPort, maxPort)
	}
	return nil
}

func (o ServiceOpts) address() string {
	return fmt.Sprintf(":%d", o.Port)
}

// ProfilerService is the data structure representing a profiler webserver.
type ProfilerService struct {
	opts   ServiceOpts
	server *http.Server
	stopFn context.CancelFunc

	log  func(format string, a ...interface{})
	warn func(err error, format string, a ...interface{})
}

// NewService returns a new Profiler instance.
func NewService(opts ServiceOpts) (*ProfilerService, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	server := &http.Server{Addr: opts.address()}
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("profiler: %s", format)
		log.Debugf(format, a...)
	}
	warnFn := func(err error, format string, a ...interface{}) {
		format = fmt.Sprintf("profiler: %s", format)
		log.WithError(err).Warnf(format, a...)
	}
	return &ProfilerService{opts, server, nil, logFn, warnFn}, nil
}

// Start starts the profiler.
func (s *ProfilerService) Start() error {
	runtime.SetBlockProfileRate(1)
	go s.server.ListenAndServe()
	ctx, cancelStats := context.WithCancel(context.Background())
	s.enableMemoryStatistics(ctx, s.opts.StatsInterval, s.opts.Datadir)
	s.stopFn = cancelStats
	s.log("start at url http://localhost:%d/debug/pprof/", s.opts.Port)
	return nil
}

// Stop stops the profiler.
func (s *ProfilerService) Stop() {
	s.stopFn()
	s.server.Shutdown(context.Background())
	s.log("stop")
}

// enableMemoryStatistics starts a goroutine that periodically logs memory
// usage of the go process to stdout.
func (s *ProfilerService) enableMemoryStatistics(
	ctx context.Context,
	interval time.Duration,
	path string,
) {

	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.printMemoryStatistics()
				s.printNumOfRoutines()
			case <-ctx.Done():
				if err := s.dumpPrometheusDefaults(path); err != nil {
					s.warn(err, "error while dumping Prometheus defaults")
				}
				return
			}
		}
	}()
}

// printMemoryStatistics logs memory statistics to stdout.
func (s *ProfilerService) printMemoryStatistics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	bytesTotalAllocated := memStats.TotalAlloc
	bytesHeapAllocated := memStats.HeapAlloc
	countMalloc := memStats.Mallocs
	countFrees := memStats.Frees

	s.log(
		"total allocated: %.3fGB, heap allocated: %.3fGB, "+
			"allocated objects count: %v, freed objects count: %v",
		toGigabytes(bytesTotalAllocated),
		toGigabytes(bytesHeapAllocated),
		countMalloc,
		countFrees,
	)
}

// printNumOfRoutines logs on stdout the number of go routines currently
// running.
func (s *ProfilerService) printNumOfRoutines() {
	log.Debugf("profiler: num of go routines: %v\n", runtime.NumGoroutine())
}

// dumpPrometheusDefaults writes default Prometheus metrics to the given file
// path.
func (s *ProfilerService) dumpPrometheusDefaults(path string) error {
	file, err := os.OpenFile(
		filepath.Join(
			path,
			time.Now().Format(time.RFC3339)),
		os.O_APPEND|os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	metricFamily, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return err
	}
	for _, v := range metricFamily {
		if _, err := writer.WriteString(v.String() + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// toGigabytes returns given memory in bytes to gigabytes.
func toGigabytes(bytes uint64) float64 {
	return float64(bytes) / gigabyte
}
