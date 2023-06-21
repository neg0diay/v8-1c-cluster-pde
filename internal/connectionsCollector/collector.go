package connectionsCollector

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

//type ConnectionInfo struct {
//	UUID        uuid.UUID `rac:"connection" json:"uuid" example:"cd16cde9-6372-4664-ac61-b0ae5cb24478"`
//	ID          int       `rac:"conn-id" json:"conn_id" example:"8714"`
//	Host        string    `json:"host" example:"host"`
//	Process     uuid.UUID `json:"process" example:"94232f94-be78-4acd-a11e-09911bd4f4ed"`
//	ClusterID   uuid.UUID `rac:"-" json:"cluster_id" example:"94232f94-be78-4acd-a11e-09911bd4f4ed"`
//	InfobaseID  uuid.UUID `json:"infobase_id" example:"efa3672f-947a-4d84-bd58-b21997b83561"`
//	Application string    `json:"application" example:"1CV8"`
//	ConnectedAt time.Time `json:"connected_at" example:"2020-10-01T07:29:57"`
//	SessionID   int       `rac:"session-number" json:"session_id" example:"148542"`
//	BlockedByLs int       `json:"blocked_by_ls" example:"0"`
//}

type connectionsCollector struct {
	ctx     context.Context
	clsuser string
	clspass string
	wg      sync.WaitGroup
	rasapi  rascli.Api

	BlockedByLs *prometheus.Desc

	Duration    *prometheus.Desc
	Connections *prometheus.Desc
}

type opt func(*connectionsCollector)

func WithCredentionals(clsuser, clspass string) opt {
	return func(c *connectionsCollector) {
		c.clsuser = clsuser
		c.clspass = clspass
	}
}

func New(rasapi rascli.Api, opts ...opt) prometheus.Collector {

	connectionLabels := []string{
		"cluster",
		"infobase",
		"connection",
		"connectionID",
		"host",
		"app",
		"process",
		"connectedAt",
		"session",
	}

	rpc := connectionsCollector{
		ctx:    context.Background(),
		rasapi: rasapi,

		BlockedByLs: prometheus.NewDesc("connection_blocked_by_ls", "connection_blocked_by_ls", connectionLabels, nil),
		Duration: prometheus.NewDesc(
			"connection_scrape_duration",
			"the time in milliseconds it took to collect the metrics",
			nil, nil),
		Connections: prometheus.NewDesc(
			"connections_total_count",
			"count of active connections on cluster",
			[]string{"cluster"}, nil),
	}

	for _, opt := range opts {
		opt(&rpc)
	}

	return &rpc
}

func (c *connectionsCollector) Describe(ch chan<- *prometheus.Desc) {
	//ch <- c.rpHosts
	ch <- c.Connections
}

func (c *connectionsCollector) Collect(ch chan<- prometheus.Metric) {

	start := time.Now()

	сlusters, err := c.rasapi.GetClusters(c.ctx)
	if err != nil {
		log.Printf("connectionsCollector: %v", err)
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

func (c *connectionsCollector) funInCollect(ch chan<- prometheus.Metric, clusterInfo serialize.ClusterInfo) {

	var (
		connectionsCount int
	)

	connections, err := c.rasapi.GetClusterConnections(c.ctx, clusterInfo.UUID)
	if err != nil {
		log.Printf("connectionsCollector: %v", err)
	}

	connections.Each(func(connectionInfo *serialize.ConnectionShortInfo) {
		var (
			sessionLabelsVal []string = []string{
				clusterInfo.Name,                                                      //"cluster",
				connectionInfo.InfobaseID.String(),                                    //"infobase",
				connectionInfo.UUID.String(),                                          //"connection",
				fmt.Sprintf("%d", connectionInfo.ID),                                  //"connectionID",
				connectionInfo.Host,                                                   //"host",
				connectionInfo.Application,                                            //"app",
				connectionInfo.Process.String(),                                       //"process",
				fmt.Sprintf("%d", connectionInfo.SessionID),                           //"session",
				connectionInfo.ConnectedAt.In(time.UTC).Format("2006-01-02 15:04:05"), //"connectedAt",
			}
		)

		ch <- prometheus.MustNewConstMetric(c.BlockedByLs, prometheus.GaugeValue, float64(connectionInfo.BlockedByLs), sessionLabelsVal...)

		connectionsCount++
	})

	ch <- prometheus.MustNewConstMetric(
		c.Connections,
		prometheus.GaugeValue,
		float64(connectionsCount),
		clusterInfo.Name)

	c.wg.Done()
}
