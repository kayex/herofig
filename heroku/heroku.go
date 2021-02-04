package heroku

import (
	"encoding/json"
	"fmt"
	"github.com/kayex/configtool/env"
	"os/exec"
)

func Pull(app string) (map[string]string, error) {
	res, err := run(app, "config", "--json")
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

func GetValue(app, key string) (string, error) {
	res, err := run(app, "config:get", key)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func SetValue(app string, key, value string) error {
	_, err := run(app, "config:set", fmt.Sprintf("%s=%s", key, value))
	return err
}

func Set(app string, config map[string]string) error {
	_, err := run(app, "config:set", string(env.FromConfig(config, " ")))
	return err
}

func run(app, script string, args ...string) ([]byte, error) {
	args = append([]string{script}, args...)
	args = append(args, "--app", app)

	cmd := exec.Command("heroku", args...)
	stdout, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed invoking Heroku CLI (%w): %s\n", err, string(ee.Stderr))
		}
		return nil, fmt.Errorf("failed invoking Heroku CLI: %w\n", err)
	}
	return stdout, err
}
