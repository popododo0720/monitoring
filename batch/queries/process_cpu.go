package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const processCpuQuery = "Process_Instance_All_CPU"

func InsertProcessCPUData(db *sql.DB, startTime, endTime time.Time, ch chan any) {
	data, err := fetchPrometheusData(processCpuQuery, startTime, endTime)
	if err != nil {
		ch <- fmt.Errorf("error fetching data from Prometheus: %w", err)
		return
	}

	stmt, err := db.Prepare("INSERT INTO instance_process_cpu(command, pid, instance_user, instance, timestamp, cpu_usage) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		ch <- fmt.Errorf("error preparing statement: %w", err)
		return
	}
	defer stmt.Close()

	for _, item := range data["result"].([]interface{}) {
		metric, ok := item.(map[string]interface{})
		if !ok {
			ch <- fmt.Errorf("unexpected result format")
			return
		}

		metricData, ok := metric["metric"].(map[string]interface{})
		if !ok {
			ch <- fmt.Errorf("unexprcted metric format")
			return
		}

		valueData, ok := metric["values"].([]interface{})
		if !ok {
			ch <- fmt.Errorf("unexpected value format")
			return
		}

		if len(metricData) < 6 {
			ch <- fmt.Errorf("metric array length is less than expected")
			return
		}

		command, ok := metricData["COMMAND"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected command format")
			return
		}

		pidStr, ok := metricData["PID"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected PID format")
			return
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			ch <- fmt.Errorf("error parsing PID: %w", err)
			return
		}

		user, ok := metricData["User"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected user format")
			return
		}

		instance, ok := metricData["instance"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected instance format")
			return
		}

		for _, v := range valueData {
			valueArray, ok := v.([]interface{})
			if !ok {
				ch <- fmt.Errorf("unexpected value array format")
				return
			}

			if len(valueArray) < 2 {
				ch <- fmt.Errorf("value array length is less than expected")
				return
			}

			var timestamp time.Time
			switch t := valueArray[0].(type) {
			case float64:
				timestamp = time.Unix(int64(t), 0).UTC()
			case string:
				ts, err := strconv.ParseFloat(t, 64)
				if err != nil {
					ch <- fmt.Errorf("error parsing timestamp: %w", err)
					return
				}
				timestamp = time.Unix(int64(ts), 0).UTC()
			default:
				ch <- fmt.Errorf("unexpected timestamp format")
				return
			}

			cpuUsageStr, ok := valueArray[1].(string)
			if !ok {
				ch <- fmt.Errorf("unexpected CPU Usage format")
				return
			}
			cpuUsage, _ := strconv.ParseFloat(cpuUsageStr, 64)

			_, err = stmt.Exec(command, pid, user, instance, timestamp, cpuUsage)
			if err != nil {
				ch <- fmt.Errorf("error inserting data into database: %w", err)
				return
			}
		}
	}

	ch <- "Successfully inserted Process CPU usage data"
	return
}

