package puller_multi_collector

import (
	"context"
	"fmt"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/puller"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PullerMultiConnector struct {
	collector []prometheus.Collector
	expose    string
}

type PullerOption func(*PullerMultiConnector)

func WithConfig(config puller.PullerConfig) PullerOption {
	return func(p *PullerMultiConnector) {
		p.expose = config.PULL_EXPOSE

	}
}

func New(collector []prometheus.Collector, opts ...PullerOption) *PullerMultiConnector {
	p := &PullerMultiConnector{
		collector: collector,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func withPanic(fn func(w http.ResponseWriter, req *http.Request)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Panic intercepted!", err)
				os.Exit(1)
			}
		}()
		fn(w, req)
	}
}

func (p *PullerMultiConnector) Run(ctx context.Context, errchan chan<- error) {

	promRegistry := prometheus.NewRegistry()
	for _, collector := range p.collector {
		promRegistry.MustRegister(collector)
	}
	//promRegistry.MustRegister(p.collector)

	handler := promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{
		//MaxRequestsInFlight: 1,
		ErrorHandling: promhttp.ContinueOnError,
		Registry:      promRegistry,
	})

	mux := http.NewServeMux()
	mux.Handle("/metrics",
		handler,
	)
	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%s", "", p.expose),
		Handler: mux,
	}

	go func() {
		errchan <- srv.ListenAndServe()
		//errchan <- http.ListenAndServe(fmt.Sprintf("%s:%s", "", p.expose), handler)
	}()
	log.Printf("v8-1c-cluster-pde: puller listen %v", fmt.Sprintf("%s:%s", "", p.expose))

	<-ctx.Done()

	if err := srv.Shutdown(context.Background()); err != nil {
		errchan <- fmt.Errorf("v8-1c-cluster-pde: puller server shutdown with err: %v", err)
	}
	log.Printf("v8-1c-cluster-pde: puller server shutdown")
}
