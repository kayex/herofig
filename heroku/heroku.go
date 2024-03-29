package heroku

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kayex/herofig/config"
	"github.com/kayex/herofig/env"
	"os/exec"
)

type Heroku struct {
	app string
}

func NewHeroku(app string) *Heroku {
	return &Heroku{app}
}

func (h *Heroku) Config() (config.Config, error) {
	res, err := h.run("config", "--json")
	if err != nil {
		return nil, err
	}

	cfg := make(map[string]string)
	err = json.Unmarshal(res, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling config JSON: %v", err)
	}
	return cfg, nil
}

func (h *Heroku) ConfigValue(key string) (string, error) {
	res, err := h.run("config:get", key)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (h *Heroku) SetConfigValue(key, value string) error {
	_, err := h.run("config:set", env.KeyValue(key, value))
	return err
}

func (h *Heroku) SetConfig(cfg config.Config) error {
	var vars []string
	for k, v := range cfg {
		vars = append(vars, env.KeyValue(k, v))
	}

	_, err := h.run("config:set", vars...)
	return err
}

func (h *Heroku) App() string {
	if h.app == "" {
		return "heroku"
	}
	return h.app
}

func (h *Heroku) run(script string, args ...string) ([]byte, error) {
	args = append([]string{script}, args...)
	if h.app != "" {
		args = append(args, "--app", h.app)
	}

	cmd := exec.Command("heroku", args...)
	stdout, err := cmd.Output()
	if err != nil {
		var e *exec.ExitError
		if errors.As(err, &e) {
			return nil, fmt.Errorf("Heroku CLI (%w): %s\n", err, string(e.Stderr))
		}
		return nil, fmt.Errorf("Heroku CLI: %w\n", err)
	}
	return stdout, err
}
