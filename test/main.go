package main

import (
    // "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strings"
)

const openrcFile = "/etc/kolla/admin-openrc"
const outputFile = "/monitoring/serverIp.json"
const apiUrl = "https://192.168.0.60:8774/v2.1/servers/detail"

func main() {
    // admin-openrc 파일을 소스
    if _, err := os.Stat(openrcFile); os.IsNotExist(err) {
        log.Fatalf("Error: %s not found", openrcFile)
    }

    // OpenStack 환경 설정 로드
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

    // Token ID 가져오기
    tokenCmd := exec.Command("openstack", "token", "issue", "-f", "value", "-c", "id")
    tokenOut, err := tokenCmd.Output()
    if err != nil {
        log.Fatalf("Failed to get token: %v", err)
    }
    tokenID := strings.TrimSpace(string(tokenOut))

    // OpenStack API 호출
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

    // JSON 데이터 파싱
    var data map[string]interface{}
    if err := json.Unmarshal(body, &data); err != nil {
        log.Fatalf("Failed to parse JSON: %v", err)
    }

    var result []map[string]string
    servers, _ := data["servers"].([]interface{})
    for _, s := range servers {
        server := s.(map[string]interface{})
        name := server["name"].(string)
        addresses := server["addresses"].(map[string]interface{})

        if internalNet, ok := addresses["internal-net"]; ok {
            for _, addr := range internalNet.([]interface{}) {
                ip := addr.(map[string]interface{})["addr"].(string)
                result = append(result, map[string]string{"name": name, "ip": ip})
            }
        }
        if externalNet, ok := addresses["external-net"]; ok {
            for _, addr := range externalNet.([]interface{}) {
                ip := addr.(map[string]interface{})["addr"].(string)
                result = append(result, map[string]string{"name": name, "ip": ip})
            }
        }
    }

    // JSON 파일로 저장
    output, err := json.MarshalIndent(result, "", "    ")
    if err != nil {
        log.Fatalf("Failed to marshal JSON: %v", err)
    }

    if err := ioutil.WriteFile(outputFile, output, 0644); err != nil {
        log.Fatalf("Failed to write JSON to file: %v", err)
    }

    fmt.Printf("Output saved to %s\n", outputFile)
}

// parseEnv 함수는 환경 변수를 파싱하는 데 사용됩니다.
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
