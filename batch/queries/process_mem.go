package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const processMemoryQuery = "Process_Instance_All_MEM"

func InsertProcessMEMData(db *sql.DB, startTime, endTime time.Time) error {
	data, err := fetchPrometheusData(processMemoryQuery, startTime, endTime)
	if err != nil {
		return fmt.Errorf("error fetching data from Prometheus: %w", err)
	}

	stmt, err := db.Prepare("INSERT INTO instance_process_mem(command, pid, instance_user, instance, timestamp, mem_usage) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range data["result"].([]interface{}) {
		metric, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected result format")
		}

		metricData, ok := metric["metric"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexprcted metric format")
		}

		valueData, ok := metric["values"].([]interface{})
		if !ok {
			return fmt.Errorf("unexpected value format")
		}

		if len(metricData) < 6 {
			return fmt.Errorf("metric array length is less than expected")
		}

		command, ok := metricData["COMMAND"].(string)
		if !ok {
			return fmt.Errorf("unexpected command format")
		}

		pidStr, ok := metricData["PID"].(string)
		if !ok {
			return fmt.Errorf("unexpected PID format")
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("error parsing PID: %w", err)
		}

		user, ok := metricData["User"].(string)
		if !ok {
			return fmt.Errorf("unexpected user format")
		}

		instance, ok := metricData["instance"].(string)
		if !ok {
			return fmt.Errorf("unexpected instance format")
		}

		for _, v := range valueData {
			valueArray, ok := v.([]interface{})
			if !ok {
				return fmt.Errorf("unexpected value array format")
			}

			if len(valueArray) < 2 {
				return fmt.Errorf("value array length is less than expected")
			}

			var timestamp time.Time
			switch t := valueArray[0].(type) {
			case float64:
				timestamp = time.Unix(int64(t), 0).UTC()
			case string:
				ts, err := strconv.ParseFloat(t, 64)
				if err != nil {
					return fmt.Errorf("error parsing timestamp: %w", err)
				}
				timestamp = time.Unix(int64(ts), 0).UTC()
			default:
				return fmt.Errorf("unexpected timestamp format")
			}

			memUsageStr, ok := valueArray[1].(string)
			if !ok {
				return fmt.Errorf("unexpected CPU Usage format")
			}
			memUsage, _ := strconv.ParseFloat(memUsageStr, 64)

			queryLog := fmt.Sprintf("INSERT INTO instance_process_mem(command, pid, instance_user, instance, timestamp, mem_usage) VALUES (%q, %d, %q, %q, %s, %f)", command, pid, user, instance, timestamp.Format(time.RFC3339), memUsage)
			fmt.Println(queryLog)

			_, err = stmt.Exec(command, pid, user, instance, timestamp, memUsage)
			if err != nil {
				return fmt.Errorf("error inserting data into database: %w", err)
			}
		}
	}

	return nil
}