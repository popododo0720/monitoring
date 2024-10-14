package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const cpuQuery = "Instance_Metrics_CPU_Usage"

func InsertCPUUsageData(db *sql.DB, startTime, endTime time.Time) error {
	data, err := fetchPrometheusData(cpuQuery, startTime, endTime)
	if err != nil {
		return fmt.Errorf("error fetching data from Prometheus: %w", err)
	}

	for _, item := range data["result"].([]interface{}) {
		metric, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected result format")
		}

		metricData, ok := metric["values"].([]interface{})
		if !ok {
			return fmt.Errorf("unexpected metric data format")
		}

		for _, value := range metricData {
			valueArray, ok := value.([]interface{})
			if !ok || len(valueArray) < 2 {
				return fmt.Errorf("unexpected metric data format")
			}

			timestampFloat, ok := valueArray[0].(float64)
			if !ok {
				return fmt.Errorf("unexpected timestamp format")
			}
			timestamp := time.Unix(int64(timestampFloat), 0).UTC()

			cpuUsageStr, ok := valueArray[1].(string)
			if !ok {
				return fmt.Errorf("unexpected CPU usage format")
			}
			cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64)
			if err != nil {
				return fmt.Errorf("error parsing CPU usage: %w", err)
			}

			ipAddress := metric["metric"].(map[string]interface{})["instance"].(string)

			_, err = db.Exec(
				"INSERT INTO cpu_usage (timestamp, cpu_usage, ip_address) VALUES ($1, $2, $3)",
				timestamp,
				cpuUsage,
				ipAddress,
			)
			if err != nil {
				return fmt.Errorf("error inserting data into database: %w", err)
			}
		}
	}

	return nil
}
