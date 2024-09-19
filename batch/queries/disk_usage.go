package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const (
	diskSizeQuery = "Instance_Metrics_DISK_Size"
	diskUsedQuery = "Instance_Metrics_DISK_Used"
)

// InsertDiskUsageData inserts disk usage data into the database
func InsertDiskUsageData(db *sql.DB, startTime, endTime time.Time) error {
	// Fetch disk size data
	sizeData, err := fetchPrometheusData(diskSizeQuery, startTime, endTime)
	if err != nil {
		return fmt.Errorf("error fetching disk size data from Prometheus: %w", err)
	}

	// Fetch disk used data
	usedData, err := fetchPrometheusData(diskUsedQuery, startTime, endTime)
	if err != nil {
		return fmt.Errorf("error fetching disk used data from Prometheus: %w", err)
	}

	for _, item := range sizeData["result"].([]interface{}) {
		metric, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected disk size result format")
		}

		metricData, ok := metric["values"].([]interface{})
		if !ok {
			return fmt.Errorf("unexpected disk size metric data format")
		}

		for _, value := range metricData {
			valueArray, ok := value.([]interface{})
			if !ok || len(valueArray) < 2 {
				return fmt.Errorf("unexpected disk size metric data format")
			}

			timestampFloat, ok := valueArray[0].(float64)
			if !ok {
				return fmt.Errorf("unexpected timestamp format")
			}
			timestamp := time.Unix(int64(timestampFloat), 0).UTC()

			diskSizeStr, ok := valueArray[1].(string)
			if !ok {
				return fmt.Errorf("unexpected disk size format")
			}
			diskSize, err := strconv.ParseFloat(diskSizeStr, 64)
			if err != nil {
				return fmt.Errorf("error parsing disk size: %w", err)
			}

			ipAddress := metric["metric"].(map[string]interface{})["instance"].(string)

			// Fetch corresponding disk used value for the same timestamp
			diskUsed, err := getMetricValueAtTime(usedData, ipAddress, timestamp)
			if err != nil {
				return fmt.Errorf("error getting disk used value: %w", err)
			}

			_, err = db.Exec(
				"INSERT INTO disk_usage (timestamp, disk_size, disk_used, ip_address) VALUES ($1, $2, $3, $4)",
				timestamp,
				int(diskSize),
				int(diskUsed),
				ipAddress,
			)
			if err != nil {
				return fmt.Errorf("error inserting data into database: %w", err)
			}
		}
	}

	return nil
}

// getMetricValueAtTime fetches the metric value for the specified IP address at a specific timestamp
func getMetricValueAtTime(data map[string]interface{}, ipAddress string, timestamp time.Time) (float64, error) {
	for _, item := range data["result"].([]interface{}) {
		metric, ok := item.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("unexpected result format")
		}

		if metric["metric"].(map[string]interface{})["instance"].(string) == ipAddress {
			metricData, ok := metric["values"].([]interface{})
			if !ok {
				return 0, fmt.Errorf("unexpected metric data format")
			}

			for _, value := range metricData {
				valueArray, ok := value.([]interface{})
				if !ok || len(valueArray) < 2 {
					return 0, fmt.Errorf("unexpected metric data format")
				}

				timestampFloat, ok := valueArray[0].(float64)
				if !ok {
					return 0, fmt.Errorf("unexpected timestamp format")
				}
				if time.Unix(int64(timestampFloat), 0).UTC() == timestamp {
					valueStr, ok := valueArray[1].(string)
					if !ok {
						return 0, fmt.Errorf("unexpected value format")
					}
					return strconv.ParseFloat(valueStr, 64)
				}
			}
		}
	}
	return 0, fmt.Errorf("no matching metric data found for IP address %s at timestamp %s", ipAddress, timestamp)
}
