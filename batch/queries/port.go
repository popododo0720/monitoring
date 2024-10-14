package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const portQuery = "Process_Instance_All_Port"

func InsertPortUsageData(db *sql.DB, startTime, endTime time.Time) error {
	data, err := fetchPrometheusData(portQuery, startTime, endTime)
	if err != nil {
		return fmt.Errorf("error fetching data from Prometheus: %w", err)
	}

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

		state, ok := metricData["State"].(string)
		if !ok {
			return fmt.Errorf("unexpected State format")
		}

		recvQ, ok := metricData["RecvQ"].(string)
		if !ok {
			return fmt.Errorf("unexpected RecvQ format")
		}

		sendQ, ok := metricData["SendQ"].(string)
		if !ok {
			return fmt.Errorf("unexpected SendQ format")
		}

		local, ok := metricData["Local"].(string)
		if !ok {
			return fmt.Errorf("unexpected Local format")
		}

		peer, ok := metricData["Peer"].(string)
		if !ok {
			return fmt.Errorf("unexpected Peer format")
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
				return fmt.Errorf("unexpected value array format")
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

			_, err = db.Exec(
				"INSERT INTO instance_ports(state, recvq, sendq, local, peer, process, timestamp, instance) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				state,
				recvQ,
				sendQ,
				local,
				peer,
				process,
				timestamp,
				instance,
			)
			if err != nil {
				return fmt.Errorf("error inserting data into database: %w", err)
			}
		}

	}

	return nil
}
