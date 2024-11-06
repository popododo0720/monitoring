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

func InsertDiskUsageData(db *sql.DB, startTime, endTime time.Time, ch chan any) {
	sizeData, err := fetchPrometheusData(diskSizeQuery, startTime, endTime)
	if err != nil {
		ch <- fmt.Sprintf("error fetching disk size data from Prometheus: %w", err)
		return
	}

	usedData, err := fetchPrometheusData(diskUsedQuery, startTime, endTime)
	if err != nil {
		ch <-fmt.Sprintf("error fetching disk used data from Prometheus: %w", err)
		return
	}

	for _, item := range sizeData["result"].([]interface{}) {
		metric, ok := item.(map[string]interface{})
		if !ok {
			ch <- fmt.Errorf("unexpected disk size result format")
			return
		}

		metricData, ok := metric["values"].([]interface{})
		if !ok {
			ch <- fmt.Errorf("unexpected disk size metric data format")
			return
		}

		for _, value := range metricData {
			valueArray, ok := value.([]interface{})
			if !ok || len(valueArray) < 2 {
				ch <- fmt.Errorf("unexpected disk size metric data format")
				return
			}

			timestampFloat, ok := valueArray[0].(float64)
			if !ok {
				ch <- fmt.Errorf("unexpected timestamp format")
				return
			}
			timestamp := time.Unix(int64(timestampFloat), 0).UTC()

			diskSizeStr, ok := valueArray[1].(string)
			if !ok {
				ch <- fmt.Errorf("unexpected disk size format")
				return
			}
			diskSize, err := strconv.ParseFloat(diskSizeStr, 64)
			if err != nil {
				ch <- fmt.Errorf("error parsing disk size: %w", err)
				return
			}

			ipAddress := metric["metric"].(map[string]interface{})["instance"].(string)

			diskUsed, err := getMetricValueAtTime(usedData, ipAddress, timestamp)
			if err != nil {
				ch <- fmt.Errorf("error getting disk used value: %w", err)
				return
			}

			_, err = db.Exec(
				"INSERT INTO disk_usage (timestamp, disk_size, disk_used, ip_address) VALUES ($1, $2, $3, $4)",
				timestamp,
				int(diskSize),
				int(diskUsed),
				ipAddress,
			)
			if err != nil {
				ch <- fmt.Errorf("error inserting data into database: %w", err)
				return
			}
		}
	}

	ch <- "Successfully inserted DISK usage data"
}

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
