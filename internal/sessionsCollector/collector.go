package sessionsCollector

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

//type SessionInfo struct {
//	ClusterID    uuid.UUID `json:"cluster_id" example:"0e588a25-8354-4344-b935-53442312aa30"`
//	UUID         uuid.UUID `rac:"session" json:"uuid" example:"1fb5f037-99e8-4924-a99d-a9e687522d32"`
//	ID           int       `rac:"session-id" json:"id" example:"12"`
//	InfobaseID   uuid.UUID `json:"infobase_id" example:"aea71760-15b3-485a-9a35-506eb8a0b04a"`
//	ConnectionID uuid.UUID `json:"connection_id" example:"8adf4514-0379-4333-a153-0b2689edc415"`
//	ProcessID    uuid.UUID `json:"process_id" example:"1af2e54f-d95a-4370-9b45-8277280cad23"`
//	UserName     string    `json:"user_name" example:"АКузнецов"`
//	Host         string    `json:"host" example:"host"`
//	AppId        string    `json:"app_id" example:"Designer"`
//	Locale       string    `json:"locale" example:"ru_RU"`
//	StartedAt    time.Time `json:"started_at" example:"2018-04-09T14:51:31"`
//
//	LastActiveAt                  time.Time `json:"last_active_at" example:"2018-04-09T14:51:31"`
//	Hibernate                     bool      `json:"hibernate" example:"true"`
//	PassiveSessionHibernateTime   int       `json:"passive_session_hibernate_time" example:"1200"`
//	HibernateDessionTerminateTime int       `json:"hibernate_dession_terminate_time" example:"86400"`
//	BlockedByDbms                 int       `json:"blocked_by_dbms" example:"0"`
//	BlockedByLs                   int       `json:"blocked_by_ls" example:"0"`
//	BytesAll                      int64     `json:"bytes_all" example:"105972550"`
//	BytesLast5min                 int64     `rac:"bytes-last-5min" json:"bytes_last_5_min" example:"0"`
//	CallsAll                      int       `json:"calls_all" example:"119052"`
//	CallsLast5min                 int64     `rac:"calls-last-5min" json:"calls_last_5_min" example:"0"`
//	DbmsBytesAll                  int64     `json:"dbms_bytes_all" example:"317824922"`
//	DbmsBytesLast5min             int64     `rac:"dbms-bytes-last-5min" json:"dbms_bytes_last_5_min" example:"0"`
//	DbProcInfo                    string    `json:"db_proc_info" example:"DbProcInfo"`
//	DbProcTook                    int       `json:"db_proc_took" example:"0"`
//	DbProcTookAt                  time.Time `json:"db_proc_took_at" example:"2018-04-09T14:51:31"`
//	DurationAll                   int       `json:"duration_all" example:"66184"`
//	DurationAllDbms               int       `json:"duration_all_dbms" example:"43242"`
//	DurationCurrent               int       `json:"duration_current" example:"0"`
//	DurationCurrentDbms           int       `json:"duration_current_dbms" example:"0"`
//	DurationLast5Min              int64     `rac:"duration-last-5min" json:"duration_last_5_min" example:"0"`
//	DurationLast5MinDbms          int64     `rac:"duration-last-5min-dbms" json:"duration_last_5_min_dbms" example:"0"`
//	MemoryCurrent                 int64     `json:"memory_current" example:"0"`
//	MemoryLast5min                int64     `json:"memory_last_5_min" example:"416379"`
//	MemoryTotal                   int64     `json:"memory_total" example:"23178863"`
//	ReadCurrent                   int64     `json:"read_current" example:"156162"`
//	ReadLast5min                  int64     `json:"read_last_5_min" example:"156162"`
//	ReadTotal                     int64     `json:"read_total" example:"15616"`
//	WriteCurrent                  int64     `json:"write_current" example:"0"`
//	WriteLast5min                 int64     `json:"write_last_5_min" example:"123"`
//	WriteTotal                    int64     `json:"write_total" example:"1071457"`
//	DurationCurrentService        int       `json:"duration_current_service" example:"0"`
//	DurationLast5minService       int64     `json:"duration_last_5_min_service" example:"30"`
//	DurationAllService            int       `json:"duration_all_service" example:"515"`
//	CurrentServiceName            string    `json:"current_service_name" example:"name"`
//	CpuTimeCurrent                int64     `json:"cpu_time_current" example:"0"`
//	CpuTimeLast5min               int64     `json:"cpu_time_last_5_min" example:"280"`
//	CpuTimeTotal                  int64     `json:"cpu_time_total" example:"6832"`
//	DataSeparation                string    `json:"data_separation" example:"sep=1"`
//	ClientIPAddress               string    `json:"client_ip_address" example:"127.0.0.1"`
//
//	Licenses *LicenseInfoList `json:"licenses"`
//}

type sessionsCollector struct {
	ctx     context.Context
	clsuser string
	clspass string
	wg      sync.WaitGroup
	rasapi  rascli.Api

	LastActiveAt                  *prometheus.Desc //datetime
	Hibernate                     *prometheus.Desc //0-1
	PassiveSessionHibernateTime   *prometheus.Desc
	HibernateDessionTerminateTime *prometheus.Desc
	BlockedByDbms                 *prometheus.Desc
	BlockedByLs                   *prometheus.Desc
	BytesAll                      *prometheus.Desc
	BytesLast5min                 *prometheus.Desc
	CallsAll                      *prometheus.Desc
	CallsLast5min                 *prometheus.Desc
	DbmsBytesAll                  *prometheus.Desc
	DbmsBytesLast5min             *prometheus.Desc
	DbProcInfo                    *prometheus.Desc // str
	DbProcTook                    *prometheus.Desc
	DbProcTookAt                  *prometheus.Desc
	DurationAll                   *prometheus.Desc
	DurationAllDbms               *prometheus.Desc
	DurationCurrent               *prometheus.Desc
	DurationCurrentDbms           *prometheus.Desc
	DurationLast5Min              *prometheus.Desc
	DurationLast5MinDbms          *prometheus.Desc
	MemoryCurrent                 *prometheus.Desc
	MemoryLast5min                *prometheus.Desc
	MemoryTotal                   *prometheus.Desc
	ReadCurrent                   *prometheus.Desc
	ReadLast5min                  *prometheus.Desc
	ReadTotal                     *prometheus.Desc
	WriteCurrent                  *prometheus.Desc
	WriteLast5min                 *prometheus.Desc
	WriteTotal                    *prometheus.Desc
	DurationCurrentService        *prometheus.Desc
	DurationLast5minService       *prometheus.Desc
	DurationAllService            *prometheus.Desc
	CurrentServiceName            *prometheus.Desc // str
	CpuTimeCurrent                *prometheus.Desc
	CpuTimeLast5min               *prometheus.Desc
	CpuTimeTotal                  *prometheus.Desc
	DataSeparation                *prometheus.Desc
	ClientIPAddress               *prometheus.Desc // str

	Duration *prometheus.Desc
	Sessions *prometheus.Desc

	//rpHosts             *prometheus.Desc
	//Duration            *prometheus.Desc
	//MemorySize          *prometheus.Desc
	//Connections         *prometheus.Desc
	//AvgThreads          *prometheus.Desc
	//AvailablePerfomance *prometheus.Desc
	//Capacity            *prometheus.Desc
	//MemoryExcessTime    *prometheus.Desc
	//SelectionSize       *prometheus.Desc
	//AvgBackCallTime     *prometheus.Desc
	//AvgCallTime         *prometheus.Desc
	//AvgDbCallTime       *prometheus.Desc
	//AvgLockCallTime     *prometheus.Desc
	//AvgServerCallTime   *prometheus.Desc
	//Running             *prometheus.Desc
	//Enable              *prometheus.Desc
}

type opt func(*sessionsCollector)

func WithCredentionals(clsuser, clspass string) opt {
	return func(c *sessionsCollector) {
		c.clsuser = clsuser
		c.clspass = clspass
	}
}

func New(rasapi rascli.Api, opts ...opt) prometheus.Collector {

	//sessionLabels := []string{"cluster", "pid", "host", "port", "startedAt"}

	sessionLabels := []string{"cluster", "session", "sessionID", "infobase", "connection", "process", "userName", "host", "app", "locale", "startedAt"}

	rpc := sessionsCollector{
		ctx:    context.Background(),
		rasapi: rasapi,

		LastActiveAt:                  prometheus.NewDesc("session_last_active_at", "session_last_active_at", sessionLabels, nil),
		Hibernate:                     prometheus.NewDesc("session_hibernate", "session_hibernate", sessionLabels, nil),
		PassiveSessionHibernateTime:   prometheus.NewDesc("session_passive_session_hibernate_time", "session_passive_session_hibernate_time", sessionLabels, nil),
		HibernateDessionTerminateTime: prometheus.NewDesc("session_hibernate_dession_terminate_time", "session_hibernate_dession_terminate_time", sessionLabels, nil),
		BlockedByDbms:                 prometheus.NewDesc("session_blocked_by_dbms", "session_blocked_by_dbms", sessionLabels, nil),
		BlockedByLs:                   prometheus.NewDesc("session_blocked_by_ls", "session_blocked_by_ls", sessionLabels, nil),
		BytesAll:                      prometheus.NewDesc("session_bytes_all", "session_bytes_all", sessionLabels, nil),
		BytesLast5min:                 prometheus.NewDesc("session_bytes_last_5min", "session_bytes_last_5min", sessionLabels, nil),
		CallsAll:                      prometheus.NewDesc("session_calls_all", "session_calls_all", sessionLabels, nil),
		CallsLast5min:                 prometheus.NewDesc("session_calls_last_5min", "session_calls_last_5min", sessionLabels, nil),
		DbmsBytesAll:                  prometheus.NewDesc("session_dbms_bytes_all", "session_dbms_bytes_all", sessionLabels, nil),
		DbmsBytesLast5min:             prometheus.NewDesc("session_dbms_last_5min", "session_dbms_last_5min", sessionLabels, nil),
		DbProcInfo:                    prometheus.NewDesc("session_db_proc_info", "session_db_proc_info", sessionLabels, nil),
		DbProcTook:                    prometheus.NewDesc("session_db_proc_took", "session_db_proc_took", sessionLabels, nil),
		DbProcTookAt:                  prometheus.NewDesc("session_db_proc_took_at", "session_db_proc_took_at", sessionLabels, nil),
		DurationAll:                   prometheus.NewDesc("session_duration_all", "session_duration_all", sessionLabels, nil),
		DurationAllDbms:               prometheus.NewDesc("session_duration_all_dbms", "session_duration_all_dbms", sessionLabels, nil),
		DurationCurrent:               prometheus.NewDesc("session_duration_current", "session_duration_current", sessionLabels, nil),
		DurationCurrentDbms:           prometheus.NewDesc("session_duration_current_dbms", "session_duration_current_dbms", sessionLabels, nil),
		DurationLast5Min:              prometheus.NewDesc("session_duration_last_5min", "session_duration_last_5min", sessionLabels, nil),
		DurationLast5MinDbms:          prometheus.NewDesc("session_duration_last_5min_dbms", "session_duration_last_5min_dbms", sessionLabels, nil),
		MemoryCurrent:                 prometheus.NewDesc("session_memory_current", "session_memory_current", sessionLabels, nil),
		MemoryLast5min:                prometheus.NewDesc("session_memory_last_5_min", "session_memory_last_5_min", sessionLabels, nil),
		MemoryTotal:                   prometheus.NewDesc("session_memory_total", "session_memory_total", sessionLabels, nil),
		ReadCurrent:                   prometheus.NewDesc("session_read_current", "session_read_current", sessionLabels, nil),
		ReadLast5min:                  prometheus.NewDesc("session_read_last_5_min", "session_read_last_5_min", sessionLabels, nil),
		ReadTotal:                     prometheus.NewDesc("session_read_total", "session_read_total", sessionLabels, nil),
		WriteCurrent:                  prometheus.NewDesc("session_write_current", "session_write_current", sessionLabels, nil),
		WriteLast5min:                 prometheus.NewDesc("session_write_last_5_min", "session_write_last_5_min", sessionLabels, nil),
		WriteTotal:                    prometheus.NewDesc("session_write_total", "session_write_total", sessionLabels, nil),
		DurationCurrentService:        prometheus.NewDesc("session_duration_current_service", "session_duration_current_service", sessionLabels, nil),
		DurationLast5minService:       prometheus.NewDesc("session_duration_last_5_min_service", "session_duration_last_5_min_service", sessionLabels, nil),
		DurationAllService:            prometheus.NewDesc("session_duration_all_service", "session_duration_all_service", sessionLabels, nil),
		CurrentServiceName:            prometheus.NewDesc("session_current_service_name", "session_current_service_name", sessionLabels, nil),
		CpuTimeCurrent:                prometheus.NewDesc("session_cpu_time_current", "session_cpu_time_current", sessionLabels, nil),
		CpuTimeLast5min:               prometheus.NewDesc("session_cpu_time_last_5_min", "session_cpu_time_last_5_min", sessionLabels, nil),
		CpuTimeTotal:                  prometheus.NewDesc("session_cpu_time_total", "session_cpu_time_total", sessionLabels, nil),
		DataSeparation:                prometheus.NewDesc("session_data_separation", "session_data_separation", sessionLabels, nil),
		ClientIPAddress:               prometheus.NewDesc("session_client_ip_address", "session_client_ip_address", sessionLabels, nil),
		Duration: prometheus.NewDesc(
			"session_scrape_duration",
			"the time in milliseconds it took to collect the metrics",
			nil, nil),
		Sessions: prometheus.NewDesc(
			"session_total_count",
			"count of active rp hosts on cluster",
			[]string{"cluster"}, nil),

		//
		//
		//rpHosts: prometheus.NewDesc(
		//	"rp_hosts_active",
		//	"count of active rp hosts on cluster",
		//	[]string{"cluster"}, nil),
		//MemorySize: prometheus.NewDesc(
		//	"rp_hosts_memory",
		//	"count of active rp hosts on cluster",
		//	proccesLabels, nil),
		//Duration: prometheus.NewDesc(
		//	"rp_hosts_scrape_duration",
		//	"the time in milliseconds it took to collect the metrics",
		//	nil, nil),
		//Connections: prometheus.NewDesc(
		//	"rp_hosts_connections",
		//	"number of connections to host",
		//	proccesLabels, nil),
		//AvgThreads: prometheus.NewDesc(
		//	"rp_hosts_avg_threads",
		//	"average number of client threads",
		//	proccesLabels, nil),
		//AvailablePerfomance: prometheus.NewDesc(
		//	"rp_hosts_available_perfomance",
		//	"available host performance",
		//	proccesLabels, nil),
		//Capacity: prometheus.NewDesc(
		//	"rp_hosts_capacity",
		//	"host capacity",
		//	proccesLabels, nil),
		//MemoryExcessTime: prometheus.NewDesc(
		//	"rp_hosts_memory_excess_time",
		//	"host memory excess time",
		//	proccesLabels, nil),
		//SelectionSize: prometheus.NewDesc(
		//	"rp_hosts_selection_size",
		//	"host selection size",
		//	proccesLabels, nil),
		//AvgBackCallTime: prometheus.NewDesc(
		//	"rp_hosts_avg_back_call_time",
		//	"host avg back call time",
		//	proccesLabels, nil),
		//AvgCallTime: prometheus.NewDesc(
		//	"rp_hosts_avg_call_time",
		//	"host avg call time",
		//	proccesLabels, nil),
		//AvgDbCallTime: prometheus.NewDesc(
		//	"rp_hosts_avg_db_call_time",
		//	"host avg db call time",
		//	proccesLabels, nil),
		//AvgLockCallTime: prometheus.NewDesc(
		//	"rp_hosts_avg_lock_call_time",
		//	"host avg lock call time",
		//	proccesLabels, nil),
		//AvgServerCallTime: prometheus.NewDesc(
		//	"rp_hosts_avg_server_call_time",
		//	"host avg server call time",
		//	proccesLabels, nil),
		//Enable: prometheus.NewDesc(
		//	"rp_hosts_enable",
		//	"host enable",
		//	proccesLabels, nil),
		//Running: prometheus.NewDesc(
		//	"rp_hosts_running",
		//	"host enable",
		//	proccesLabels, nil),
	}

	for _, opt := range opts {
		opt(&rpc)
	}

	return &rpc
}

func (c *sessionsCollector) Describe(ch chan<- *prometheus.Desc) {
	//ch <- c.rpHosts
	ch <- c.Sessions
}

func (c *sessionsCollector) Collect(ch chan<- prometheus.Metric) {

	start := time.Now()

	сlusters, err := c.rasapi.GetClusters(c.ctx)
	if err != nil {
		log.Printf("sessionsCollector: %v", err)
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

func (c *sessionsCollector) funInCollect(ch chan<- prometheus.Metric, clusterInfo serialize.ClusterInfo) {

	var (
		sessionsCount int
	)

	sessions, err := c.rasapi.GetClusterSessions(c.ctx, clusterInfo.UUID)
	if err != nil {
		log.Printf("sessionsCollector: %v", err)
	}

	sessions.Each(func(sessionInfo *serialize.SessionInfo) {
		var (
			sessionLabelsVal []string = []string{
				clusterInfo.Name,                  //"cluster",
				sessionInfo.UUID.String(),         //session
				fmt.Sprintf("%d", sessionInfo.ID), //"sessionID",
				sessionInfo.InfobaseID.String(),   //"infobase",
				sessionInfo.ConnectionID.String(), //"connection",
				sessionInfo.ProcessID.String(),    //"process",
				sessionInfo.UserName,              //"userName",
				sessionInfo.Host,                  //"host",
				sessionInfo.AppId,                 //"app",
				sessionInfo.Locale,                //"locale",
				sessionInfo.StartedAt.In(time.UTC).Format("2006-01-02 15:04:05"), //"startedAt",

				//sessionInfo.DbProcInfo,
				//sessionInfo.CurrentServiceName,
				//sessionInfo.DataSeparation,
				//sessionInfo.ClientIPAddress,
			}
		)

		ch <- prometheus.MustNewConstMetric(c.LastActiveAt, prometheus.GaugeValue, float64(sessionInfo.LastActiveAt.In(time.UTC).Unix()), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.Hibernate, prometheus.GaugeValue,
			func(fl bool) float64 {
				if fl {
					return 1.0
				} else {
					return 0.0
				}
			}(sessionInfo.Hibernate), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.PassiveSessionHibernateTime, prometheus.GaugeValue, float64(sessionInfo.PassiveSessionHibernateTime), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.HibernateDessionTerminateTime, prometheus.GaugeValue, float64(sessionInfo.HibernateDessionTerminateTime), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.BlockedByDbms, prometheus.GaugeValue, float64(sessionInfo.BlockedByDbms), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.BlockedByLs, prometheus.GaugeValue, float64(sessionInfo.BlockedByLs), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.BytesAll, prometheus.GaugeValue, float64(sessionInfo.BytesAll), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.BytesLast5min, prometheus.GaugeValue, float64(sessionInfo.BytesLast5min), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.CallsAll, prometheus.GaugeValue, float64(sessionInfo.CallsAll), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.CallsLast5min, prometheus.GaugeValue, float64(sessionInfo.CallsLast5min), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DbmsBytesAll, prometheus.GaugeValue, float64(sessionInfo.DbmsBytesAll), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DbmsBytesLast5min, prometheus.GaugeValue, float64(sessionInfo.DbmsBytesLast5min), sessionLabelsVal...)
		//ch <- prometheus.MustNewConstMetric(c.DbProcInfo, prometheus.GaugeValue, sessionInfo.DbProcInfo, sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DbProcTook, prometheus.GaugeValue, float64(sessionInfo.DbProcTook), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DbProcTookAt, prometheus.GaugeValue, float64(sessionInfo.DbProcTookAt.In(time.UTC).Unix()), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationAll, prometheus.GaugeValue, float64(sessionInfo.DurationAll), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationAllDbms, prometheus.GaugeValue, float64(sessionInfo.DurationAllDbms), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationCurrent, prometheus.GaugeValue, float64(sessionInfo.DurationCurrent), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationCurrentDbms, prometheus.GaugeValue, float64(sessionInfo.DurationCurrentDbms), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationLast5Min, prometheus.GaugeValue, float64(sessionInfo.DurationLast5Min), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationLast5MinDbms, prometheus.GaugeValue, float64(sessionInfo.DurationLast5MinDbms), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.MemoryCurrent, prometheus.GaugeValue, float64(sessionInfo.MemoryCurrent), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.MemoryLast5min, prometheus.GaugeValue, float64(sessionInfo.MemoryLast5min), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.MemoryTotal, prometheus.GaugeValue, float64(sessionInfo.MemoryTotal), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.ReadCurrent, prometheus.GaugeValue, float64(sessionInfo.ReadCurrent), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.ReadLast5min, prometheus.GaugeValue, float64(sessionInfo.ReadLast5min), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.ReadTotal, prometheus.GaugeValue, float64(sessionInfo.ReadTotal), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.WriteCurrent, prometheus.GaugeValue, float64(sessionInfo.WriteCurrent), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.WriteLast5min, prometheus.GaugeValue, float64(sessionInfo.WriteLast5min), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.WriteTotal, prometheus.GaugeValue, float64(sessionInfo.WriteTotal), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationCurrentService, prometheus.GaugeValue, float64(sessionInfo.DurationCurrentService), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationLast5minService, prometheus.GaugeValue, float64(sessionInfo.DurationLast5minService), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.DurationAllService, prometheus.GaugeValue, float64(sessionInfo.DurationAllService), sessionLabelsVal...)
		//ch <- prometheus.MustNewConstMetric(c.CurrentServiceName, prometheus.GaugeValue, float64(sessionInfo.CurrentServiceName), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.CpuTimeCurrent, prometheus.GaugeValue, float64(sessionInfo.CpuTimeCurrent), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.CpuTimeLast5min, prometheus.GaugeValue, float64(sessionInfo.CpuTimeLast5min), sessionLabelsVal...)
		ch <- prometheus.MustNewConstMetric(c.CpuTimeTotal, prometheus.GaugeValue, float64(sessionInfo.CpuTimeTotal), sessionLabelsVal...)
		//ch <- prometheus.MustNewConstMetric(c.DataSeparation, prometheus.GaugeValue, float64(sessionInfo.DataSeparation), sessionLabelsVal...)
		//ch <- prometheus.MustNewConstMetric(c.ClientIPAddress, prometheus.GaugeValue, float64(sessionInfo.ClientIPAddress), sessionLabelsVal...)

		sessionsCount++
	})

	ch <- prometheus.MustNewConstMetric(
		c.Sessions,
		prometheus.GaugeValue,
		float64(sessionsCount),
		clusterInfo.Name)

	c.wg.Done()

	//workingProcesses.Each(func(proccesInfo *serialize.ProcessInfo) {
	//
	//	var (
	//		proccesLabelsVal []string = []string{
	//			clusterInfo.Name,
	//			proccesInfo.Pid,
	//			fmt.Sprint(proccesInfo.Host),
	//			fmt.Sprint(proccesInfo.Port),
	//			// lag MSK+3
	//			proccesInfo.StartedAt.In(time.UTC).Format("2006-01-02 15:04:05"),
	//		}
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.MemorySize,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.MemorySize),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.Connections,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.Connections),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.AvgThreads,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.AvgThreads),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.AvailablePerfomance,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.AvailablePerfomance),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.Capacity,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.Capacity),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.MemoryExcessTime,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.MemoryExcessTime),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.SelectionSize,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.SelectionSize),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.AvgBackCallTime,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.AvgBackCallTime),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.AvgCallTime,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.AvgCallTime),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.AvgDbCallTime,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.AvgDbCallTime),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.AvgLockCallTime,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.AvgLockCallTime),
	//		proccesLabelsVal...,
	//	)
	//
	//	ch <- prometheus.MustNewConstMetric(
	//		c.AvgServerCallTime,
	//		prometheus.GaugeValue,
	//		float64(proccesInfo.AvgServerCallTime),
	//		proccesLabelsVal...,
	//	)
	//	ch <- prometheus.MustNewConstMetric(
	//		c.Enable,
	//		prometheus.GaugeValue,
	//		func(fl bool) float64 {
	//			if fl {
	//				return 1.0
	//			} else {
	//				return 0.0
	//			}
	//		}(proccesInfo.Enable),
	//		proccesLabelsVal...,
	//	)
	//	ch <- prometheus.MustNewConstMetric(
	//		c.Running,
	//		prometheus.GaugeValue,
	//		func(fl bool) float64 {
	//			if fl {
	//				return 1.0
	//			} else {
	//				return 0.0
	//			}
	//		}(proccesInfo.Running),
	//		proccesLabelsVal...,
	//	)
	//
	//	rpHostsCount++
	//})
	//
	//ch <- prometheus.MustNewConstMetric(
	//	c.rpHosts,
	//	prometheus.GaugeValue,
	//	float64(rpHostsCount),
	//	clusterInfo.Name)
	//
	//c.wg.Done()
}
