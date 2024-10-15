package queries

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	prometheusURL = "http://10.0.2.110:9091/api/v1/query_range"
	username      = "admin"
	password      = "JOBdRRi8IvBPnmSdxZ1V7tum15VfaJcVkQ5zigZ6"
)

func fetchPrometheusData(query string, startTime, endTime time.Time) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", prometheusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("start", startTime.Format(time.RFC3339))
	q.Add("end", endTime.Format(time.RFC3339))
	q.Add("step", "60")
	req.URL.RawQuery = q.Encode()

	auth := fmt.Sprintf("%s:%s", username, password)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON response: %w", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format: %v", result)
	}

	return data, nil
}
