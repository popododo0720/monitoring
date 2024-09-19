package main

import (
	"log"
	"time"

	"main/common"
	"main/queries"
)

func main() {
	db, err := common.ConnectDB()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	// // cpu
	if err := queries.InsertCPUUsageData(db, time.Now().Add(-1*time.Hour), time.Now()); err != nil {
		log.Fatalf("Error inserting CPU usage data: %v", err)
	}

	// // mem
	if err := queries.InsertMemoryUsageData(db, time.Now().Add(-1*time.Hour), time.Now()); err != nil {
		log.Fatalf("Error inserting MEM usage data: %v", err)
	}

	// // disk
	if err := queries.InsertDiskUsageData(db, time.Now().Add(-1*time.Hour), time.Now()); err != nil {
		log.Fatalf("Error inserting Disk usage data: %v", err)
	}

	// process cpu
	if err := queries.InsertProcessCPUData(db, time.Now().Add(-1*time.Hour), time.Now()); err != nil {
		log.Fatalf("Error inserting Process CPU usage data: %v", err)
	}

	// process mem
	if err := queries.InsertProcessMEMData(db, time.Now().Add(-1*time.Hour), time.Now()); err != nil {
		log.Fatalf("Error inserting Process MEM usage data: %v", err)
	}
}
