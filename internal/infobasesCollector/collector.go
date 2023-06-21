package infobasesCollector

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"log"
	"sync"
	"time"

	rascli "github.com/khorevaa/ras-client"
	"github.com/khorevaa/ras-client/serialize"
	"github.com/prometheus/client_golang/prometheus"
)

//type InfobaseInfo struct {
//	UUID                                   uuid.UUID `rac:"infobase" json:"uuid" example:"efa3672f-947a-4d84-bd58-b21997b83561"`
//	Name                                   string    `json:"name" example:"УППБоеваяБаза"`
//	Description                            string    `rac:"descr" json:"descr" example:"Это очень хорошая база"`
//	Dbms                                   string    `json:"dbms" example:"MSSQLServer"`
//	DbServer                               string    `json:"db_server" example:"sql"`
//	DbName                                 string    `json:"db_name" example:"base"`
//	DbUser                                 string    `json:"db_user" example:"user"`
//	DbPwd                                  string    `rac:"-" json:"db_pwd" example:"password"`
//	SecurityLevel                          int       `json:"security_level" example:"0"`
//	LicenseDistribution                    int       `json:"license_distribution" example:"0"`
//	ScheduledJobsDeny                      bool      `json:"scheduled_jobs_deny" example:"false"`
//	SessionsDeny                           bool      `json:"sessions_deny" example:"false"`
//	DeniedFrom                             time.Time `json:"denied_from" example:"2020-10-01T08:30:00Z"`
//	DeniedMessage                          string    `json:"denied_message" example:"Выполняется обновление базы"`
//	DeniedParameter                        string    `json:"denied_parameter" example:"123"`
//	DeniedTo                               time.Time `json:"denied_to" example:"2020-10-01T08:30:00Z"`
//	PermissionCode                         string    `json:"permission_code" example:"123"`
//	ExternalSessionManagerConnectionString string    `json:"external_session_manager_connection_string" example:"http://auth2.com"`
//	ExternalSessionManagerRequired         bool      `json:"external_session_manager_required" example:"false"`
//	SecurityProfileName                    string    `json:"security_profile_name"  example:"sec_profile1"`
//	SafeModeSecurityProfileName            string    `json:"safe_mode_security_profile_name" example:"profile1"`
//	ReserveWorkingProcesses                bool      `json:"reserve_working_processes" example:"false"`
//	DateOffset                             int       `json:"date_offset" example:"0"`
//	Locale                                 string    `json:"locale" example:"ru_RU"`
//	ClusterID                              uuid.UUID `rac:"-" json:"cluster_id" example:"efa3672f-947a-4d84-bd58-b21997b83561"`
//
//	Cluster     *ClusterInfo             `json:"-" swaggerignore:"true"`
//	Connections *ConnectionShortInfoList `json:"-" swaggerignore:"true"`
//	Sessions    *SessionInfoList         `json:"-" swaggerignore:"true"`
//	Locks       *LocksList               `json:"-" swaggerignore:"true"`
//}

type InfobaseSummaryInfo struct {
	ClusterID   uuid.UUID `rac:"-" json:"cluster_id" example:""`
	UUID        uuid.UUID `rac:"infobase" json:"uuid" example:"efa3672f-947a-4d84-bd58-b21997b83561"`
	Name        string    `json:"name" example:"УППБоеваяБаза"`
	Description string    `rac:"descr" json:"descr" example:"УППБоеваяБаза"`
}

type infobasesCollector struct {
	ctx     context.Context
	clsuser string
	clspass string
	wg      sync.WaitGroup
	rasapi  rascli.Api

	Infobase *prometheus.Desc

	//Connections *prometheus.Desc
	//Sessions    *prometheus.Desc
	//Locks       *prometheus.Desc

	Duration  *prometheus.Desc
	Infobases *prometheus.Desc
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
		"name",
		"description",
		//"dbms",
		//"dbServer",
		//"dbName",
		//"dbUser",
		//"dbPwd",
		//"securityLevel",
		//"licenseDistribution",
		//"scheduledJobsDeny",
		//"sessionsDeny",
		//"deniedFrom",
		//"deniedMessage",
		//"deniedParameter",
		//"deniedTo",
		//"permissionCode",
		//"externalSessionManagerConnectionString",
		//"externalSessionManagerRequired",
		//"securityProfileName",
		//"safeModeSecurityProfileName",
		//"reserveWorkingProcesses",
		//"dateOffset",
		//"locale",
	}

	rpc := infobasesCollector{
		ctx:    context.Background(),
		rasapi: rasapi,

		Infobase: prometheus.NewDesc("infobase_infobase", "infobase_infobase", infobaseLabels, nil),

		//Connections: prometheus.NewDesc("infobase_connections", "infobase_connections", infobaseLabels, nil),
		//Sessions:    prometheus.NewDesc("infobase_sessions", "infobase_sessions", infobaseLabels, nil),
		//Locks:       prometheus.NewDesc("infobase_locks", "infobase_locks", infobaseLabels, nil),

		Duration: prometheus.NewDesc(
			"infobase_scrape_duration",
			"the time in milliseconds it took to collect the metrics",
			nil, nil),
		Infobases: prometheus.NewDesc(
			"infobases_total_count",
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
	ch <- c.Infobases
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
		infobasesCount int
	)

	infobases, err := c.rasapi.GetClusterInfobases(c.ctx, clusterInfo.UUID)
	if err != nil {
		log.Printf("infobasesCollector: %v", err)
	}

	infobases.Each(func(infobaseSummaryInfo *serialize.InfobaseSummaryInfo) {

		//infobaseInfo, getInfobaseInfoErr := c.rasapi.GetInfobaseInfo(c.ctx, clusterInfo.UUID, infobaseSummaryInfo.UUID)
		//if err != nil {
		//	log.Printf("infobasesCollector: %v", getInfobaseInfoErr)
		//	return
		//}

		var (
			infobaseLabelsVal []string = []string{
				clusterInfo.Name,                  //"cluster",
				infobaseSummaryInfo.UUID.String(), //"infobase",
				infobaseSummaryInfo.Name,          //"name",
				infobaseSummaryInfo.Description,   //"description",
				//infobaseInfo.Dbms,                                   //"dbms",
				//infobaseInfo.DbServer,                               //"dbServer",
				//infobaseInfo.DbName,                                 //"dbName",
				//infobaseInfo.DbUser,                                 //"dbUser",
				//infobaseInfo.DbPwd,                                  //"dbPwd",
				//fmt.Sprintf("%d", infobaseInfo.SecurityLevel),       //"securityLevel",
				//fmt.Sprintf("%d", infobaseInfo.LicenseDistribution), //"licenseDistribution",
				//fmt.Sprintf("%t", infobaseInfo.ScheduledJobsDeny),   //"scheduledJobsDeny",
				//fmt.Sprintf("%t", infobaseInfo.SessionsDeny),        //"sessionsDeny",
				//infobaseInfo.DeniedFrom.In(time.UTC).Format("2006-01-02 15:04:05"), //"deniedFrom",
				//infobaseInfo.DeniedMessage,                                       //"deniedMessage",
				//infobaseInfo.DeniedParameter,                                     //"deniedParameter",
				//infobaseInfo.DeniedTo.In(time.UTC).Format("2006-01-02 15:04:05"), //"deniedTo",
				//infobaseInfo.PermissionCode,                                      //"permissionCode",
				//infobaseInfo.ExternalSessionManagerConnectionString,              //"externalSessionManagerConnectionString",
				//fmt.Sprintf("%t", infobaseInfo.ExternalSessionManagerRequired),   //"externalSessionManagerRequired",
				//infobaseInfo.SecurityProfileName,                                 //"securityProfileName",
				//infobaseInfo.SafeModeSecurityProfileName,                         //"safeModeSecurityProfileName",
				//fmt.Sprintf("%t", infobaseInfo.ReserveWorkingProcesses),          //"reserveWorkingProcesses",
				//fmt.Sprintf("%d", infobaseInfo.DateOffset),                       //"dateOffset",
				//infobaseInfo.Locale,                                              //"locale",
			}
		)

		//ch <- prometheus.MustNewConstMetric(c.Sessions, prometheus.GaugeValue, float64(len(*infobaseInfo.Sessions)), infobaseLabelsVal...)
		//ch <- prometheus.MustNewConstMetric(c.Connections, prometheus.GaugeValue, float64(len(*infobaseInfo.Connections)), infobaseLabelsVal...)
		//ch <- prometheus.MustNewConstMetric(c.Locks, prometheus.GaugeValue, float64(len(*infobaseInfo.Locks)), infobaseLabelsVal...)

		ch <- prometheus.MustNewConstMetric(c.Infobase, prometheus.GaugeValue, 1, infobaseLabelsVal...)

		//ch <- prometheus.MustNewConstMetric(c.Sessions, prometheus.GaugeValue, 0, infobaseLabelsVal...)
		//ch <- prometheus.MustNewConstMetric(c.Connections, prometheus.GaugeValue, 0, infobaseLabelsVal...)
		//ch <- prometheus.MustNewConstMetric(c.Locks, prometheus.GaugeValue, 0, infobaseLabelsVal...)

		infobasesCount++
	})

	ch <- prometheus.MustNewConstMetric(
		c.Infobases,
		prometheus.GaugeValue,
		float64(infobasesCount),
		clusterInfo.Name)

	c.wg.Done()
}
