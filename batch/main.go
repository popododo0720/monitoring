package main

import (
	"log"
	"time"
	"os"

	"main/common"
	"main/queries"
)

func main() {

	currentDate := time.Now().Format("2006-01-02")

	logFilePath := "/monitoring/batch/batchlogs/batch_process_" + currentDate + ".log"

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	action := os.Args[1]

	log.Println("Start Batch Process")

	db, err := common.ConnectDB()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	switch action {
	case "cpu":
		if err := queries.InsertCPUUsageData(db, time.Now().Add(-1 * time.Hour), time.Now()); err != nil {
			log.Fatalf("Error inserting CPU usage data: %v", err)
		} else {
			log.Println("Successfully inserted CPU usage data")
		}

	case "mem":
		if err := queries.InsertMemoryUsageData(db, time.Now().Add(-1 * time.Hour), time.Now()); err != nil {
			log.Fatalf("Error inserting MEM usage data: %v", err)
		} else {
			log.Println("Successfully inserted MEM usage data")
		}

	case "disk":
		if err := queries.InsertDiskUsageData(db, time.Now().Add(-1 * time.Hour), time.Now()); err != nil {
			log.Fatalf("Error inserting Disk usage data: %v", err)
		} else {
			log.Println("Successfully inserted Disk usage data")
		}

	case "process-cpu":
		if err := queries.InsertProcessCPUData(db, time.Now().Add(-10 * time.Minute), time.Now()); err != nil {
			log.Fatalf("Error inserting Process CPU usage data: %v", err)
		} else {
			log.Println("Successfully inserted Process CPU usage data")
		}

	case "process-mem":
		if err := queries.InsertProcessMEMData(db, time.Now().Add(-10 * time.Minute), time.Now()); err != nil {
			log.Fatalf("Error inserting Process MEM usage data: %v", err)
		} else {
			log.Println("Successfully inserted Process MEM usage data")
		}

	case "instance-port":
		if err := queries.InsertPortUsageData(db, time.Now().Add(-10 * time.Minute), time.Now()); err != nil {
			log.Fatalf("Error inserting Process All Port data: %v", err)
		} else {
			log.Println("Successfully inserted Port usage data")
		}
	default:
		log.Printf("Unknown action: %s", action)
	}
}
