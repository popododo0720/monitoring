package queries

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/lib/pq"
)

const (
	openrcFile = "/etc/kolla/admin-openrc"
	apiUrl     = "https://192.168.0.60:8774/v2.1/servers/detail"
)

func InsertReportMaping(db *sql.DB, ch chan any) {
	if _, err := os.Stat(openrcFile); os.IsNotExist(err) {
		log.Fatalf("Error: %s not found", openrcFile)
	}

	sourceCmd := fmt.Sprintf("source %s && env", openrcFile)
	cmd := exec.Command("bash", "-c", sourceCmd)
	envOut, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to source openrc: %v", err)
	}

	envVars := parseEnv(string(envOut))
	for k, v := range envVars {
		os.Setenv(k, v)
	}

	tokenCmd := exec.Command("openstack", "token", "issue", "-f", "value", "-c", "id")
	tokenOut, err := tokenCmd.Output()
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	tokenID := strings.TrimSpace(string(tokenOut))

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Add("X-Auth-Token", tokenID)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get instance details: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS report_mapping")
	if err != nil {
		log.Fatalf("Failed to drop table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE report_mapping (
			id SERIAL PRIMARY KEY,
			display_name TEXT UNIQUE,
			ip_address TEXT[],
			report_type TEXT
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	servers, _ := data["servers"].([]interface{})

	for _, s := range servers {
		server := s.(map[string]interface{})
		name := server["name"].(string)
		addresses := server["addresses"].(map[string]interface{})

		var ipAddresses []string
		for _, netAddresses := range addresses {
			for _, addr := range netAddresses.([]interface{}) {
				ip := addr.(map[string]interface{})["addr"].(string)
				ipAddresses = append(ipAddresses, ip)
			}
		}

		_, err := db.Exec(
			"INSERT INTO report_mapping (display_name, ip_address) VALUES ($1, $2) ON CONFLICT (display_name) DO UPDATE SET ip_address = EXCLUDED.ip_address",
			name,
			pq.Array(ipAddresses),
		)
		if err != nil {
			ch <- fmt.Errorf("error inserting data into database: %w", err)
			return
		}
	}

	ch <- "Successfully inserted Report Mapping data"
}

func parseEnv(envOutput string) map[string]string {
	envVars := make(map[string]string)
	lines := strings.Split(envOutput, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}
	return envVars
}