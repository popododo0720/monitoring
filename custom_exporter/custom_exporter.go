package main

import (
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"main/instance_metrics"
	"main/process"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "customExporter"
)

var (
	processLabelNames = []string{"custom_500k"}
)

type metricInfo struct {
	Desc *prometheus.Desc
	Type prometheus.ValueType
}

func newProcessMetric(metricName string, docString string, t prometheus.ValueType, constLabels prometheus.Labels) metricInfo {
	return metricInfo{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "custom_500k", metricName),
			docString,
			processLabelNames,
			constLabels,
		),
		Type: t,
	}
}

type metrics map[int]metricInfo

func (m metrics) String() string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	s := make([]string, len(keys))
	for i, k := range keys {
		s[i] = strconv.Itoa(k)
	}
	return strings.Join(s, ",")
}

var (
	processMetrics = metrics{
		2: newProcessMetric("Instance_Metrics_CPU_Usage", "Instance_Metrics_CPU_Usage.", prometheus.GaugeValue, nil),
		3: newProcessMetric("Instance_Metrics_MEM_Usage", "Instance_Metrics_MEM_Usage.", prometheus.GaugeValue, nil),

		4: newProcessMetric("Instance_Metrics_DISK_Size", "Instance_Metrics_DISK_Size.", prometheus.GaugeValue, nil),
		5: newProcessMetric("Instance_Metrics_DISK_Used", "Instance_Metrics_DISK_Used.", prometheus.GaugeValue, nil),
		6: newProcessMetric("Instance_Metrics_DISK_Avail", "Instance_Metrics_DISK_Avail.", prometheus.GaugeValue, nil),
		7: newProcessMetric("Instance_Metrics_DISK_UseRate", "Instance_Metrics_DISK_UseRate.", prometheus.GaugeValue, nil),

		8: newProcessMetric("Instance_Metrics_All_CPU_Usage", "Instance_Metrics_All_CPU_Usage.", prometheus.GaugeValue, nil),
		9: newProcessMetric("Instance_Metrics_All_MEM_Usage", "Instance_Metrics_All_MEM_Usage.", prometheus.GaugeValue, nil),

		10: newProcessMetric("Instance_Metrics_All_Port", "Instance_Metrics_All_Port.", prometheus.GaugeValue, nil),

		11: newProcessMetric("Instance_Metrics_Disk_IOWait", "Instance_Metrics_Disk_IOWait.", prometheus.GaugeValue, nil),
		12: newProcessMetric("Instance_Metrics_Disk_BI", "Instance_Metrics_Disk_BI.", prometheus.GaugeValue, nil),
		13: newProcessMetric("Instance_Metrics_Disk_BO", "Instance_Metrics_Disk_BO.", prometheus.GaugeValue, nil),
	}
)

type Exporter struct {
	mutex                         sync.RWMutex
	Instance_Metrics_CPU_Usage    prometheus.Gauge
	Instance_Metrics_MEM_Usage    prometheus.Gauge
	Instance_Metrics_DISK_Size    prometheus.Gauge
	Instance_Metrics_DISK_Used    prometheus.Gauge
	Instance_Metrics_DISK_Avail   prometheus.Gauge
	Instance_Metrics_DISK_UseRate prometheus.Gauge
	Process_Instance_All_CPU      *prometheus.GaugeVec
	Process_Instance_All_MEM      *prometheus.GaugeVec
	Process_Instance_All_Port     *prometheus.GaugeVec
	Instance_Metrics_Disk_IOWait  prometheus.Gauge
	Instance_Metrics_Disk_BI      prometheus.Gauge
	Instance_Metrics_Disk_BO      prometheus.Gauge
}

func NewExporter() *Exporter {
	return &Exporter{
		Instance_Metrics_CPU_Usage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_CPU_Usage",
			Help: "Openstack Instance CPU Usage",
		}),
		Instance_Metrics_MEM_Usage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_MEM_Usage",
			Help: "Openstack Instance MEM Usage",
		}),
		Instance_Metrics_DISK_Size: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_DISK_Size",
			Help: "Openstack Instance DISK Size",
		}),
		Instance_Metrics_DISK_Used: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_DISK_Used",
			Help: "Openstack Instance DISK Used",
		}),
		Instance_Metrics_DISK_Avail: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_DISK_Avail",
			Help: "Openstack Instance DISK Avail",
		}),
		Instance_Metrics_DISK_UseRate: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_DISK_UseRate",
			Help: "Openstack Instance DISK UseRate",
		}),
		Process_Instance_All_CPU: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "Process_Instance_All_CPU",
			Help: "CPU Usage of each process",
		}, []string{"User", "PID", "COMMAND"}),
		Process_Instance_All_MEM: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "Process_Instance_All_MEM",
			Help: "Memory Usage of each process",
		}, []string{"User", "PID", "COMMAND"}),
		Process_Instance_All_Port: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "Process_Instance_All_Port",
			Help: "Instance All Port List",
		}, []string{"State", "RecvQ", "SendQ", "Local", "Peer", "Process"}),
		Instance_Metrics_Disk_IOWait: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_Dist_IOWait",
			Help: "Openstack Instance Disk IOWait",
		}),
		Instance_Metrics_Disk_BI: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_Disk_BI",
			Help: "Openstack Instance Disk BI",
		}),
		Instance_Metrics_Disk_BO: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "Instance_Metrics_Disk_BO",
			Help: "Openstack Instance Disk BO",
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.Instance_Metrics_CPU_Usage.Describe(ch)
	e.Instance_Metrics_MEM_Usage.Describe(ch)
	e.Instance_Metrics_DISK_Size.Describe(ch)
	e.Instance_Metrics_DISK_Used.Describe(ch)
	e.Instance_Metrics_DISK_Avail.Describe(ch)
	e.Instance_Metrics_DISK_UseRate.Describe(ch)
	e.Process_Instance_All_CPU.Describe(ch)
	e.Process_Instance_All_MEM.Describe(ch)
	e.Process_Instance_All_Port.Describe(ch)
	e.Instance_Metrics_Disk_IOWait.Describe(ch)
	e.Instance_Metrics_Disk_BI.Describe(ch)
	e.Instance_Metrics_Disk_BO.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.Process_Instance_All_CPU.Reset()
	e.Process_Instance_All_MEM.Reset()
	e.Process_Instance_All_Port.Reset()

	cpuUsage, err := instance_metrics.GetInstanceCpuUsage()
	if err != nil {
		fmt.Printf("Error getting CPU usage: %v\n", err)
		return
	}
	memUsage, err := instance_metrics.GetInstanceMemUsage()
	if err != nil {
		fmt.Printf("Error getting CPU usage: %v\n", err)
		return
	}
	diskSize, err := instance_metrics.GetInstanceDiskSize()
	if err != nil {
		fmt.Printf("Error getting CPU usage: %v\n", err)
		return
	}
	diskUsed, err := instance_metrics.GetInstanceDiskUsed()
	if err != nil {
		fmt.Printf("Error getting CPU usage: %v\n", err)
		return
	}
	diskAvail, err := instance_metrics.GetInstanceDiskAvail()
	if err != nil {
		fmt.Printf("Error getting CPU usage: %v\n", err)
		return
	}
	diskUseRate, err := instance_metrics.GetInstanceDiskUseRate()
	if err != nil {
		fmt.Printf("Error getting CPU usage: %v\n", err)
		return
	}
	processList, err := process.GetProcessList()
	if err != nil {
		fmt.Printf("Error getting Instance Process List: %v\n", err)
		return
	}
	portList, err := process.GetPortList()
	if err != nil {
		fmt.Printf("Error getting Instance Port List: %v\n", err)
		return
	}
	iowait, err := process.GetInstanceIOWait()
	if err != nil {
		fmt.Printf("Error getting Instance IOWait: %v\n", err)
		return
	}
	diskbi, err := process.GetInstanceBi()
	if err != nil {
		fmt.Printf("Error getting Instance BI: %v\n", err)
		return
	}
	diskbo, err := process.GetInstanceBo()
	if err != nil {
		fmt.Printf("Error getting Instance BO: %v\n", err)
		return
	}

	e.Instance_Metrics_CPU_Usage.Set(cpuUsage)
	e.Instance_Metrics_MEM_Usage.Set(memUsage)
	e.Instance_Metrics_DISK_Size.Set(diskSize)
	e.Instance_Metrics_DISK_Used.Set(diskUsed)
	e.Instance_Metrics_DISK_Avail.Set(diskAvail)
	e.Instance_Metrics_DISK_UseRate.Set(diskUseRate)
	for _, proc := range processList {
		e.Process_Instance_All_CPU.WithLabelValues(
			proc.User,
			proc.PID,
			proc.COMMAND,
		).Set(proc.CPU)

		e.Process_Instance_All_MEM.WithLabelValues(
			proc.User,
			proc.PID,
			proc.COMMAND,
		).Set(proc.MEM)
	}
	for _, port := range portList {
		e.Process_Instance_All_Port.WithLabelValues(
			port.State,
			port.RecvQ,
			port.SendQ,
			port.Local,
			port.Peer,
			port.Process,
		).Set(1)
	}
	e.Instance_Metrics_Disk_IOWait.Set(iowait)
	e.Instance_Metrics_Disk_BI.Set(diskbi)
	e.Instance_Metrics_Disk_BO.Set(diskbo)

	e.Instance_Metrics_CPU_Usage.Collect(ch)
	e.Instance_Metrics_MEM_Usage.Collect(ch)
	e.Instance_Metrics_DISK_Size.Collect(ch)
	e.Instance_Metrics_DISK_Used.Collect(ch)
	e.Instance_Metrics_DISK_Avail.Collect(ch)
	e.Instance_Metrics_DISK_UseRate.Collect(ch)
	e.Process_Instance_All_CPU.Collect(ch)
	e.Process_Instance_All_MEM.Collect(ch)
	e.Process_Instance_All_Port.Collect(ch)
	e.Instance_Metrics_Disk_IOWait.Collect(ch)
	e.Instance_Metrics_Disk_BI.Collect(ch)
	e.Instance_Metrics_Disk_BO.Collect(ch)
}

func main() {
	port := flag.String("web.listen-address", ":8088", "Address on which to expose metrics and web interface.")
	flag.Parse()

	exporter := NewExporter()
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Printf("Listening on %s\n", *port)
	if err := http.ListenAndServe(*port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

//
