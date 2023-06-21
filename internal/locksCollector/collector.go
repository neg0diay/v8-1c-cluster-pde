package locksCollector

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

//type LockInfo struct {
//	ClusterID    uuid.UUID `json:"cluster_id" example:"00000000-0000-0000-0000-000000000000"`
//	InfobaseID   uuid.UUID `json:"infobase_id" example:"8b8a0817-4cb1-4e13-9a8f-472dde1a3b47"`
//	ConnectionID uuid.UUID `json:"connection_id" example:"00000000-0000-0000-0000-000000000000"`
//	SessionID    uuid.UUID `json:"session_id" example:"8b8a0817-4cb1-4e13-9a8f-472dde1a3b47"`
//	ObjectID     uuid.UUID `json:"object_id" example:"00000000-0000-0000-0000-000000000000"`
//	LockedAt     time.Time `json:"locked_at" example:"2020-10-01T08:30:00"`
//	Description  string    `rac:"descr" json:"descr" example:"БД(сеанс ,УППБоеваяБаза,разделяемая)"`
//}

type infobasesCollector struct {
	ctx     context.Context
	clsuser string
	clspass string
	wg      sync.WaitGroup
	rasapi  rascli.Api

	Lock *prometheus.Desc

	//Connections *prometheus.Desc
	//Sessions    *prometheus.Desc
	//Locks       *prometheus.Desc

	Duration *prometheus.Desc
	Locks    *prometheus.Desc
}

type opt func(*infobasesCollector)

func WithCredentionals(clsuser, clspass string) opt {
	return func(c *infobasesCollector) {
		c.clsuser = clsuser
		c.clspass = clspass
	}
}

func New(rasapi rascli.Api, opts ...opt) prometheus.Collector {

	infobaseLabels := []string{
		"cluster",
		"infobase",
		"connection",
		"session",
		"object",
		"lockedAt",
		"description",
	}

	rpc := infobasesCollector{
		ctx:    context.Background(),
		rasapi: rasapi,

		Lock: prometheus.NewDesc("lock_lock", "lock_lock", infobaseLabels, nil),

		//Connections: prometheus.NewDesc("infobase_connections", "infobase_connections", infobaseLabels, nil),
		//Sessions:    prometheus.NewDesc("infobase_sessions", "infobase_sessions", infobaseLabels, nil),
		//Locks:       prometheus.NewDesc("infobase_locks", "infobase_locks", infobaseLabels, nil),

		Duration: prometheus.NewDesc(
			"lock_scrape_duration",
			"the time in milliseconds it took to collect the metrics",
			nil, nil),
		Locks: prometheus.NewDesc(
			"lock_total_count",
			"count of active infobases on cluster",
			[]string{"cluster"}, nil),
	}

	for _, opt := range opts {
		opt(&rpc)
	}

	return &rpc
}

func (c *infobasesCollector) Describe(ch chan<- *prometheus.Desc) {
	//ch <- c.rpHosts
	ch <- c.Locks
}

func (c *infobasesCollector) Collect(ch chan<- prometheus.Metric) {

	start := time.Now()

	сlusters, err := c.rasapi.GetClusters(c.ctx)
	if err != nil {
		log.Printf("infobasesCollector: %v", err)
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

func (c *infobasesCollector) funInCollect(ch chan<- prometheus.Metric, clusterInfo serialize.ClusterInfo) {

	var (
		locksCount int
	)

	locks, err := c.rasapi.GetClusterLocks(c.ctx, clusterInfo.UUID)
	if err != nil {
		log.Printf("locksCollector: %v", err)
	}

	locks_added := make(map[string]int)

	locks.Each(func(lockInfo *serialize.LockInfo) {

		var (
			lockLabelsVal []string = []string{

				clusterInfo.Name,                                             //"cluster",
				lockInfo.InfobaseID.String(),                                 //"infobase",
				lockInfo.ConnectionID.String(),                               //"connection",
				lockInfo.SessionID.String(),                                  //"session",
				lockInfo.ObjectID.String(),                                   //"object",
				lockInfo.LockedAt.In(time.UTC).Format("2006-01-02 15:04:05"), //"lockedAt",
				lockInfo.Description,                                         //"description",
			}
		)

		concat_kyes :=
			lockInfo.InfobaseID.String() +
				lockInfo.ConnectionID.String() +
				lockInfo.SessionID.String() +
				lockInfo.ObjectID.String() +
				lockInfo.LockedAt.In(time.UTC).Format("2006-01-02 15:04:05") +
				lockInfo.Description

		v, ok := locks_added[concat_kyes]

		if !ok {
			ch <- prometheus.MustNewConstMetric(c.Lock, prometheus.GaugeValue, 1, lockLabelsVal...)
			locks_added[concat_kyes] = 1
		} else {
			fmt.Println(v)
		}

		locksCount++
	})

	ch <- prometheus.MustNewConstMetric(
		c.Locks,
		prometheus.GaugeValue,
		float64(locksCount),
		clusterInfo.Name)

	c.wg.Done()
}
