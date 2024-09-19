package process

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func GetInstanceIOWait() (float64, error) {
	cmd := exec.Command("sh", "-c", "vmstat | tail -n 1 | awk '{print $16}'")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("error executing command: %w", err)
	}

	output := strings.TrimSpace(out.String())

	ioWait, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting output to float: %w", err)
	}

	return ioWait, nil
}
