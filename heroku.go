package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type Heroku struct {
	app string
}

func NewHeroku(app string) *Heroku {
	return &Heroku{app}
}

func (h *Heroku) Config() (Config, error) {
	res, err := h.run("config", "--json")
	if err != nil {
		return nil, err
	}

	cfg := make(map[string]string)
	err = json.Unmarshal(res, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling config JSON: %v", err)
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
	_, err := h.run("config:set", Var{key, value}.String())
	return err
}

func (h *Heroku) SetConfig(cfg Config) error {
	var vars []string
	for k, v := range cfg {
		vars = append(vars, Var{k, v}.String())
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
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("Heroku CLI (%w): %s\n", err, string(ee.Stderr))
		}
		return nil, fmt.Errorf("Heroku CLI: %w\n", err)
	}
	return stdout, err
}

func (h *Heroku) authenticated() (bool, error) {
	cmd := exec.Command("heroku", "whoami")
	err := cmd.Run()
	var ee *exec.ExitError
	if err != nil {
		if errors.As(err, &ee) {
			if ee.ExitCode() == 100 {
				return false, nil
			}
			return false, fmt.Errorf("Heroku CLI (%w): %s\n", err, string(ee.Stderr))
		}
	}
	return true, nil
}
