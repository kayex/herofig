package heroku

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func Pull(app string) (map[string]string, error) {
	res, err := run(app, "config", []string{"--json"})
	if err != nil {
		return nil, err
	}

	config := make(map[string]string)
	err = json.Unmarshal([]byte(res), &config)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling config JSON: %v", err)
	}
	return config, nil
}

func run(app, script string, args []string) ([]byte, error) {
	cmd := exec.Command(script, args...)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed invoking Heroku CLI: %w", err)
	}
	return stdout, err
}