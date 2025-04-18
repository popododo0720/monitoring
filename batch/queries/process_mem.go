package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

type ProcessMemoryData struct {
	Command   string
	PID       int
	User      string
	Instance  string
	Timestamp time.Time
	MemUsage  float64
}

const processMemoryQuery = "Process_Instance_All_MEM"

func InsertProcessMEMData(db *sql.DB, startTime, endTime time.Time, ch chan any) {

	data, err := fetchPrometheusData(processMemoryQuery, startTime, endTime)
	if err != nil {
		ch <- fmt.Errorf("error fetching data from Prometheus: %w", err)
		return
	}

	var processDataList []ProcessMemoryData

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

			memUsageStr, ok := valueArray[1].(string)
			if !ok {
				ch <- fmt.Errorf("unexpected CPU Usage format")
				return
			}
			memUsage, _ := strconv.ParseFloat(memUsageStr, 64)

			processDataList = append(processDataList, ProcessMemoryData{
				Command:   command,
				PID:       pid,
				User:      user,
				Instance:  instance,
				Timestamp: timestamp,
				MemUsage:  memUsage,
			})

		}
	}

	err = insertProcessMemDataDB(db, processDataList)
	if err != nil {
		ch <- fmt.Errorf("error inserting data into database: %w", err)
		return
	}

	ch <- "Successfully inserted Process MEM usage data"
}

func insertProcessMemDataDB(db *sql.DB, processDataList []ProcessMemoryData) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transation: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO instance_process_mem(command, pid, instance_user, instance, timestamp, mem_usage) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	batchSize := 10000
	for i := 0; i < len(processDataList); i += batchSize {
		end := i + batchSize
		if end > len(processDataList) {
			end = len(processDataList)
		}

		for _, data := range processDataList[i:end] {
			_, err = stmt.Exec(data.Command, data.PID, data.User, data.Instance, data.Timestamp, data.MemUsage)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error inserting data into database: %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
