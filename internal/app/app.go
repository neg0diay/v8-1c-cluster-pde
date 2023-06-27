package app

import (
	"context"
	"flag"
	"fmt"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/clustersCollector"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/collector"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/connectionsCollector"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/infobasesCollector"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/locksCollector"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/sessionsCollector"
	//"github.com/prometheus/client_golang/prometheus"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Chipazawra/v8-1c-cluster-pde/internal/puller"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/puller_multi_collector"
	pusher "github.com/Chipazawra/v8-1c-cluster-pde/internal/pusher"
	"github.com/Chipazawra/v8-1c-cluster-pde/internal/rpHostsCollector"
	"github.com/caarlos0/env"
	rascli "github.com/khorevaa/ras-client"
)

const (
	push string = "push"
	pull string = "pull"
)

var (
	conf          AppConfig
	RAS_HOST      string
	RAS_PORT      string
	RAS_VERSION   string
	MODE          string
	PULL_EXPOSE   string
	PUSH_INTERVAL int
	PUSH_HOST     string
	PUSH_PORT     string
	AGNT_PASS     string
	AGNT_USER     string
	CLS_USER      string
	CLS_PASS      string
)

func init() {

	if err := env.Parse(&conf); err != nil {
		log.Fatalf("app: config...")
	}

	log.Printf("v8-1c-cluster-pde: config from env:\n %#v", conf)

	flag.StringVar(&RAS_HOST, "ras-host", "", "cluster host.")
	flag.StringVar(&RAS_PORT, "ras-port", "", "cluster port.")
	flag.StringVar(&RAS_VERSION, "ras-version", "", "cluster version.")

	flag.StringVar(&PULL_EXPOSE, "pull-expose", "", "metrics port.")
	flag.StringVar(&MODE, "mode", "", "mode push or pull")
	flag.IntVar(&PUSH_INTERVAL, "push-interval", 0, "mode push or pull")
	flag.StringVar(&PUSH_HOST, "push-host", "", "pushgateway host")
	flag.StringVar(&PUSH_PORT, "push-port", "", "pushgateway port")
	flag.StringVar(&AGNT_USER, "agnt-user", "", "agent user")
	flag.StringVar(&AGNT_PASS, "agnt-pass", "", "agent password")
	flag.StringVar(&CLS_USER, "cls-user", "", "cluster user")
	flag.StringVar(&CLS_PASS, "cls-pass", "", "cluster password")
	flag.Parse()

	if RAS_HOST != "" {
		conf.RAS_HOST = RAS_HOST
	}

	if RAS_PORT != "" {
		conf.RAS_PORT = RAS_PORT
	}

	if PULL_EXPOSE != "" {
		conf.PULL_EXPOSE = PULL_EXPOSE
	}

	if MODE != "" {
		conf.MODE = MODE
	}

	if PUSH_INTERVAL != 0 {
		conf.PUSH_INTERVAL = PUSH_INTERVAL
	}

	if PUSH_HOST != "" {
		conf.PUSH_HOST = PUSH_HOST
	}

	if PUSH_PORT != "" {
		conf.PUSH_PORT = PUSH_PORT
	}

	if AGNT_USER != "" {
		conf.AGNT_USER = AGNT_USER
	}

	if AGNT_PASS != "" {
		conf.AGNT_PASS = AGNT_PASS
	}

	if CLS_USER != "" {
		conf.CLS_USER = CLS_USER
	}

	if CLS_PASS != "" {
		conf.CLS_PASS = CLS_PASS
	}

	log.Printf("v8-1c-cluster-pde: overrided config from stdin:\n%#v", conf)
}

func Run() error {

	rcli := rascli.NewClient(fmt.Sprintf("%s:%s", conf.RAS_HOST, conf.RAS_PORT), rascli.WithVersion(conf.RAS_VERSION))
	rcli.AuthenticateAgent(conf.AGNT_USER, conf.AGNT_USER)

	defer rcli.Close()

	log.Printf("v8-1c-cluster-pde: connected to RAS %v",
		fmt.Sprintf("%s:%s", conf.RAS_HOST, conf.RAS_PORT),
	)

	rhc := rpHostsCollector.New(rcli,
		rpHostsCollector.WithCredentionals(conf.CLS_USER, conf.CLS_PASS),
	)

	sc := sessionsCollector.New(rcli,
		sessionsCollector.WithCredentionals(conf.CLS_USER, conf.CLS_PASS),
	)

	cc := connectionsCollector.New(rcli,
		connectionsCollector.WithCredentionals(conf.CLS_USER, conf.CLS_PASS),
	)

	ibc := infobasesCollector.New(rcli,
		infobasesCollector.WithCredentionals(conf.CLS_USER, conf.CLS_PASS),
	)

	lc := locksCollector.New(rcli,
		locksCollector.WithCredentionals(conf.CLS_USER, conf.CLS_PASS),
	)

	clc := clustersCollector.New(rcli,
		clustersCollector.WithCredentionals(conf.CLS_USER, conf.CLS_PASS),
	)

	ctx, cancel := context.WithCancel(context.Background())

	sigchan := make(chan os.Signal, 1)
	defer close(sigchan)
	errchan := make(chan error)
	defer close(errchan)

	signal.Notify(sigchan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var collecter Collecter
	//var sc_collecter Collecter

	switch conf.MODE {
	case push:
		collecter = pusher.New(rhc, pusher.WithConfig(
			pusher.PusherConfig{
				PUSH_INTERVAL: conf.PUSH_INTERVAL,
				PUSH_HOST:     conf.PUSH_HOST,
				PUSH_PORT:     conf.PUSH_PORT,
			}))
	case pull:

		collectors := []collector.Collector{
			rhc,
			sc,
			cc,
			ibc,
			lc,
			clc,
		}

		collecter = puller_multi_collector.New(collectors, puller_multi_collector.WithConfig(
			puller.PullerConfig{
				PULL_EXPOSE: conf.PULL_EXPOSE,
			}))
	}

	log.Printf("v8-1c-cluster-pde: runing in %v mode", conf.MODE)

	go collecter.Run(ctx, errchan)

	select {
	case sig := <-sigchan:
		cancel()
		log.Printf("v8-1c-cluster-pde: received signal %v", sig)
		return nil
	case err := <-errchan:
		cancel()
		return err
	}
}

type Collecter interface {
	Run(ctx context.Context, errchan chan<- error)
}
