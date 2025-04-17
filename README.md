# Prometheus Exporter Custom

## metric
| Metric | Type |
| ------ | ------ |
| Instance_Metrics_CPU_Usage | prometheus.Gauge |
| Instance_Metrics_MEM_Usage | prometheus.Gauge |
| Instance_Metrics_DISK_Size | prometheus.Gauge |
| Instance_Metrics_DISK_Used | prometheus.Gauge |
| Instance_Metrics_DISK_Avail | prometheus.Gauge |
| Instance_Metrics_DISK_UseRate | prometheus.Gauge |
| Process_Instance_All_CPU | *prometheus.GaugeVec |
| Process_Instance_All_MEM | *prometheus.GaugeVec |
| Process_Instance_All_Port | *prometheus.GaugeVec |
| Instance_Metrics_Disk_IOWait | prometheus.Gauge |
| Instance_Metrics_Disk_BI | prometheus.Gauge |
| Instance_Metrics_Disk_BO | prometheus.Gauge |

