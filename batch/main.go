package main

import (
	"log"
	"time"
	"os"

	"main/common"
	"main/queries"
)

func main() {

	start := time.Now()

	currentDate := time.Now().Format("2006-01-02")

	logFilePath := "/monitoring/batch/batchlogs/batch_process_" + currentDate + ".log"

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("Start Batch Process")

	db, err := common.ConnectDB()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	c := make(chan any, 1)

	go queries.InsertCPUUsageData(db, time.Now().Add(-1 * time.Hour), time.Now(), c)
	go queries.InsertMemoryUsageData(db, time.Now().Add(-1 * time.Hour), time.Now(), c)
	go queries.InsertDiskUsageData(db, time.Now().Add(-1 * time.Hour), time.Now(), c)
	go queries.InsertPortUsageData(db, time.Now().Add(-1 * time.Hour), time.Now(), c)
	go queries.InsertProcessCPUData(db, time.Now().Add(-1 * time.Hour), time.Now(), c)
	go queries.InsertProcessMEMData(db, time.Now().Add(-1 * time.Hour), time.Now(), c)

	for i := 0; i < 6; i++ { 
		result := <-c 
		log.Println(result) 
	}

	log.Println("took: ", time.Since(start))

	close(c)
}
