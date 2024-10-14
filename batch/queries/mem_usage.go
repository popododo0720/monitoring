package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const memoryQuery = "Instance_Metrics_MEM_Usage"

func InsertMemoryUsageData(db *sql.DB, startTime, endTime time.Time) error {
	data, err := fetchPrometheusData(memoryQuery, startTime, endTime)
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

			memUsageStr, ok := valueArray[1].(string)
			if !ok {
				return fmt.Errorf("unexpected memory usage format")
			}
			memUsage, err := strconv.ParseFloat(memUsageStr, 64)
			if err != nil {
				return fmt.Errorf("error parsing memory usage: %w", err)
			}

			ipAddress := metric["metric"].(map[string]interface{})["instance"].(string)

			_, err = db.Exec(
				"INSERT INTO mem_usage (timestamp, mem_usage, ip_address) VALUES ($1, $2, $3)",
				timestamp,
				memUsage,
				ipAddress,
			)
			if err != nil {
				return fmt.Errorf("error inserting data into database: %w", err)
			}
		}
	}

	return nil
}
