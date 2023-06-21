package rpHostsCollector

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	rascli "github.com/khorevaa/ras-client"
	"github.com/khorevaa/ras-client/serialize"
	"github.com/prometheus/client_golang/prometheus"
)

//type ProcessInfo struct {
//	UUID                uuid.UUID       `rac:"process" json:"uuid" example:"0e588a25-8354-4344-b935-53442312aa30"`
//	Host                string          `json:"host" example:"srv"`
//	Port                int16           `json:"port" example:"1564"`
//	Pid                 string          `json:"pid" example:"3366"`
//	Enable              bool            `rac:"is-enable" json:"enable" example:"true"`
//	Running             bool            `json:"running" example:"true"`
//	StartedAt           time.Time       `json:"started_at" example:"2018-03-29T11:16:02"`
//	Use                 bool            `json:"use" example:"true"`
//
//	AvailablePerfomance int             `json:"available_perfomance" example:"100"`
//	Capacity            int             `json:"capacity" example:"1000"`
//	Connections         int             `json:"connections" example:"7"`
//	MemorySize          int             `json:"memory_size" example:"1518604"`
//	MemoryExcessTime    int             `json:"memory_excess_time" example:"0"`
//	SelectionSize       int             `json:"selection_size" example:"61341"`
//	AvgBackCallTime     float64         `json:"avg_back_call_time" example:"0.000"`
//	AvgCallTime         float64         `json:"avg_call_time" example:"0.483"`
//	AvgDbCallTime       float64         `json:"avg_db_call_time" example:"0.124"`
//	AvgLockCallTime     float64         `json:"avg_lock_call_time" example:"0.000"`
//	AvgServerCallTime   float64         `json:"avg_server_call_time" example:"-0.265"`
//	AvgThreads          float64         `json:"avg_threads" example:"0.281"`
//	Reverse             bool            `json:"reverse" example:"true"`
//	Licenses            LicenseInfoList `json:"licenses"`
//
//	ClusterID uuid.UUID `json:"cluster_id" example:"0e588a25-8354-4344-b935-53442312aa30"`
//}

type rpHostsCollector struct {
	ctx                 context.Context
	clsuser             string
	clspass             string
	wg                  sync.WaitGroup
	rasapi              rascli.Api
	rpHosts             *prometheus.Desc
	Duration            *prometheus.Desc
	MemorySize          *prometheus.Desc
	Connections         *prometheus.Desc
	AvgThreads          *prometheus.Desc
	AvailablePerfomance *prometheus.Desc
	Capacity            *prometheus.Desc
	MemoryExcessTime    *prometheus.Desc
	SelectionSize       *prometheus.Desc
	AvgBackCallTime     *prometheus.Desc
	AvgCallTime         *prometheus.Desc
	AvgDbCallTime       *prometheus.Desc
	AvgLockCallTime     *prometheus.Desc
	AvgServerCallTime   *prometheus.Desc
	Running             *prometheus.Desc
	Enable              *prometheus.Desc
}

type opt func(*rpHostsCollector)

func WithCredentionals(clsuser, clspass string) opt {
	return func(c *rpHostsCollector) {
		c.clsuser = clsuser
		c.clspass = clspass
	}
}

func New(rasapi rascli.Api, opts ...opt) prometheus.Collector {

	proccesLabels := []string{
		"cluster",
		"procces",
		"proccesID",
		"host",
		"port",
		"enable",
		"running",
		"use",
		"startedAt",
	}

	rpc := rpHostsCollector{
		ctx:    context.Background(),
		rasapi: rasapi,
		rpHosts: prometheus.NewDesc(
			"rp_hosts_active",
			"count of active rp hosts on cluster",
			[]string{"cluster"}, nil),
		MemorySize: prometheus.NewDesc(
			"rp_hosts_memory",
			"count of active rp hosts on cluster",
			proccesLabels, nil),
		Duration: prometheus.NewDesc(
			"rp_hosts_scrape_duration",
			"the time in milliseconds it took to collect the metrics",
			nil, nil),
		Connections: prometheus.NewDesc(
			"rp_hosts_connections",
			"number of connections to host",
			proccesLabels, nil),
		AvgThreads: prometheus.NewDesc(
			"rp_hosts_avg_threads",
			"average number of client threads",
			proccesLabels, nil),
		AvailablePerfomance: prometheus.NewDesc(
			"rp_hosts_available_perfomance",
			"available host performance",
			proccesLabels, nil),
		Capacity: prometheus.NewDesc(
			"rp_hosts_capacity",
			"host capacity",
			proccesLabels, nil),
		MemoryExcessTime: prometheus.NewDesc(
			"rp_hosts_memory_excess_time",
			"host memory excess time",
			proccesLabels, nil),
		SelectionSize: prometheus.NewDesc(
			"rp_hosts_selection_size",
			"host selection size",
			proccesLabels, nil),
		AvgBackCallTime: prometheus.NewDesc(
			"rp_hosts_avg_back_call_time",
			"host avg back call time",
			proccesLabels, nil),
		AvgCallTime: prometheus.NewDesc(
			"rp_hosts_avg_call_time",
			"host avg call time",
			proccesLabels, nil),
		AvgDbCallTime: prometheus.NewDesc(
			"rp_hosts_avg_db_call_time",
			"host avg db call time",
			proccesLabels, nil),
		AvgLockCallTime: prometheus.NewDesc(
			"rp_hosts_avg_lock_call_time",
			"host avg lock call time",
			proccesLabels, nil),
		AvgServerCallTime: prometheus.NewDesc(
			"rp_hosts_avg_server_call_time",
			"host avg server call time",
			proccesLabels, nil),
		Enable: prometheus.NewDesc(
			"rp_hosts_enable",
			"host enable",
			proccesLabels, nil),
		Running: prometheus.NewDesc(
			"rp_hosts_running",
			"host enable",
			proccesLabels, nil),
	}

	for _, opt := range opts {
		opt(&rpc)
	}

	return &rpc
}

func (c *rpHostsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rpHosts
}

func (c *rpHostsCollector) Collect(ch chan<- prometheus.Metric) {

	start := time.Now()

	сlusters, err := c.rasapi.GetClusters(c.ctx)
	if err != nil {
		log.Printf("rpHostsCollector: %v", err)
	}

	for _, сluster := range сlusters {
		c.wg.Add(1)
		c.rasapi.AuthenticateCluster(сluster.UUID, c.clsuser, c.clspass)
		go c.funInCollect(ch, *сluster)
	}
	c.wg.Wait()

	ch <- prometheus.MustNewConstMetric(
		c.Duration,
		prometheus.GaugeValue,
		float64(time.Since(start).Milliseconds()))
}

func (c *rpHostsCollector) funInCollect(ch chan<- prometheus.Metric, clusterInfo serialize.ClusterInfo) {

	var (
		rpHostsCount int
	)

	workingProcesses, err := c.rasapi.GetWorkingProcesses(c.ctx, clusterInfo.UUID)
	if err != nil {
		log.Printf("rpHostsCollector: %v", err)
	}

	workingProcesses.Each(func(proccesInfo *serialize.ProcessInfo) {

		var (
			proccesLabelsVal []string = []string{
				clusterInfo.Name,
				proccesInfo.UUID.String(),    //UUID
				proccesInfo.Pid,              //Pid
				fmt.Sprint(proccesInfo.Host), //Host
				fmt.Sprint(proccesInfo.Port), //Port
				fmt.Sprint(proccesInfo.Enable),
				fmt.Sprint(proccesInfo.Running),
				fmt.Sprint(proccesInfo.Use),
				// lag MSK+3
				proccesInfo.StartedAt.In(time.UTC).Format("2006-01-02 15:04:05"), //StartedAt
			}
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemorySize,
			prometheus.GaugeValue,
			float64(proccesInfo.MemorySize),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Connections,
			prometheus.GaugeValue,
			float64(proccesInfo.Connections),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgThreads,
			prometheus.GaugeValue,
			float64(proccesInfo.AvgThreads),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvailablePerfomance,
			prometheus.GaugeValue,
			float64(proccesInfo.AvailablePerfomance),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Capacity,
			prometheus.GaugeValue,
			float64(proccesInfo.Capacity),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MemoryExcessTime,
			prometheus.GaugeValue,
			float64(proccesInfo.MemoryExcessTime),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SelectionSize,
			prometheus.GaugeValue,
			float64(proccesInfo.SelectionSize),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgBackCallTime,
			prometheus.GaugeValue,
			float64(proccesInfo.AvgBackCallTime),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgCallTime,
			prometheus.GaugeValue,
			float64(proccesInfo.AvgCallTime),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgDbCallTime,
			prometheus.GaugeValue,
			float64(proccesInfo.AvgDbCallTime),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgLockCallTime,
			prometheus.GaugeValue,
			float64(proccesInfo.AvgLockCallTime),
			proccesLabelsVal...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AvgServerCallTime,
			prometheus.GaugeValue,
			float64(proccesInfo.AvgServerCallTime),
			proccesLabelsVal...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Enable,
			prometheus.GaugeValue,
			func(fl bool) float64 {
				if fl {
					return 1.0
				} else {
					return 0.0
				}
			}(proccesInfo.Enable),
			proccesLabelsVal...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Running,
			prometheus.GaugeValue,
			func(fl bool) float64 {
				if fl {
					return 1.0
				} else {
					return 0.0
				}
			}(proccesInfo.Running),
			proccesLabelsVal...,
		)

		rpHostsCount++
	})

	ch <- prometheus.MustNewConstMetric(
		c.rpHosts,
		prometheus.GaugeValue,
		float64(rpHostsCount),
		clusterInfo.Name)

	c.wg.Done()
}
