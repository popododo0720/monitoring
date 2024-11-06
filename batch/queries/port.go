package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

type ProcessPortData struct {
	State     string
	RecvQ     string
	SendQ     string
	Local     string
	Peer      string
	Process   string
	Timestamp time.Time
	Instance  string
}

const portQuery = "Process_Instance_All_Port"

func InsertPortUsageData(db *sql.DB, startTime, endTime time.Time, ch chan any) {
	data, err := fetchPrometheusData(portQuery, startTime, endTime)
	if err != nil {
		ch <- fmt.Errorf("error fetching data from Prometheus: %w", err)
		return
	}

	var processDataList []ProcessPortData

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

		state, ok := metricData["State"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected State format")
			return
		}

		recvQ, ok := metricData["RecvQ"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected RecvQ format")
			return
		}

		sendQ, ok := metricData["SendQ"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected SendQ format")
			return
		}

		local, ok := metricData["Local"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected Local format")
			return
		}

		peer, ok := metricData["Peer"].(string)
		if !ok {
			ch <- fmt.Errorf("unexpected Peer format")
			return
		}

		process, ok := metricData["Process"].(string)
		if !ok {
			process = ""
		}

		instance, ok := metricData["instance"].(string)
		if !ok {
			process = ""
		}

		for _, v := range valueData {
			valueArray, ok := v.([]interface{})
			if !ok {
				ch <- fmt.Errorf("unexpected value array format")
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

			processDataList = append(processDataList, ProcessPortData{
				State:     state,
				RecvQ:     recvQ,
				SendQ:     sendQ,
				Local:     local,
				Peer:      peer,
				Process:   process,
				Timestamp: timestamp,
				Instance:  instance,
			})

		}

	}

	err = insertPortUsageDataDB(db, processDataList, ch)
	if err != nil {
		ch <- fmt.Errorf("error inserting data into database: %w", err)
		return
	}

	ch <- "Successfully inserted Process Port usage data"
}

func insertPortUsageDataDB(db *sql.DB, processDataList []ProcessPortData, ch chan any) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO instance_ports(state, recvq, sendq, local, peer, process, timestamp, instance) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
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

		for _, data := range processDataList {
			_, err = stmt.Exec(data.State, data.RecvQ, data.SendQ, data.Local, data.Peer, data.Process, data.Timestamp, data.Instance)
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
