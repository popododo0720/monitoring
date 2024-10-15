package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const memoryQuery = "Instance_Metrics_MEM_Usage"

func InsertMemoryUsageData(db *sql.DB, startTime, endTime time.Time, ch chan any) {
	data, err := fetchPrometheusData(memoryQuery, startTime, endTime)
	if err != nil {
		ch <- fmt.Errorf("error fetching data from Prometheus: %v", err)
		return
	}

	for _, item := range data["result"].([]interface{}) {
		metric, ok := item.(map[string]interface{})
		if !ok {
			ch <- fmt.Errorf("unexpected result format")
			return
		}

		metricData, ok := metric["values"].([]interface{})
		if !ok {
			ch <- fmt.Errorf("unexpected metric data format")
			return
		}

		for _, value := range metricData {
			valueArray, ok := value.([]interface{})
			if !ok || len(valueArray) < 2 {
				ch <- fmt.Errorf("unexpected metric data format")
				return
			}

			timestampFloat, ok := valueArray[0].(float64)
			if !ok {
				ch <- fmt.Errorf("unexpected timestamp format")
				return
			}
			timestamp := time.Unix(int64(timestampFloat), 0).UTC()

			memUsageStr, ok := valueArray[1].(string)
			if !ok {
				ch <-  fmt.Errorf("unexpected memory usage format")
				return
			}
			memUsage, err := strconv.ParseFloat(memUsageStr, 64)
			if err != nil {
				ch <- fmt.Errorf("error parsing memory usage: %w", err)
				return
			}

			ipAddress := metric["metric"].(map[string]interface{})["instance"].(string)

			_, err = db.Exec(
				"INSERT INTO mem_usage (timestamp, mem_usage, ip_address) VALUES ($1, $2, $3)",
				timestamp,
				memUsage,
				ipAddress,
			)
			if err != nil {
				ch <- fmt.Errorf("error inserting data into database: %w", err)
				return
			}
		}
	}

	ch <- "Successfully inserted Mem usage data"
	return
}
