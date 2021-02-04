package heroku

import (
	"encoding/json"
	"fmt"
	"github.com/kayex/configtool/env"
	"os/exec"
)

type Heroku struct {
	app string
}

func New(app string) *Heroku {
	return &Heroku{app: app}
}

func (h *Heroku) Name() string {
	return h.app
}

func (h *Heroku) Get() (map[string]string, error) {
	res, err := run(h.app, "config", "--json")
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

func (h *Heroku) GetValue(key string) (string, error) {
	res, err := run(h.app, "config:get", key)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (h *Heroku) SetValue(key, value string) error {
	_, err := run(h.app, "config:set", fmt.Sprintf("%s=%s", key, value))
	return err
}

func (h *Heroku) Set(config map[string]string) error {
	var vars []string
	for k, v := range config {
		vars = append(vars, env.KeyValue(k, v))
	}

	_, err := run(h.app, "config:set", vars...)
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
