package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ParseVar(v string) (key string, value string, err error) {
	delimiter := strings.Index(v, "=")
	if delimiter < 1 {
		return "", "", fmt.Errorf("invalid env variable format %q", v)
	}
	pr := []rune(v)

	key = string(pr[:delimiter])
	key = strings.TrimSpace(key)
	value = string(pr[delimiter+1:])
	return key, value, nil
}

func Parse(env io.Reader) (Config, error) {
	cfg := make(Config)

	scanner := bufio.NewScanner(env)
	scanner.Split(bufio.ScanLines)

	line := 0
	for scanner.Scan() {
		line++
		t := scanner.Text()
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		k, v, err := ParseVar(t)
		if err != nil {
			return nil, fmt.Errorf("processing line %d: %v", line, err)
		}
		cfg[k] = v
	}
	return cfg, nil
}

func Load(filename string) (Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing env file: %v", err)
	}

	return cfg, nil
}

func Save(filename string, cfg Config) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	for _, v := range cfg.Ordered() {
		_, err := fmt.Fprintln(f, v.String())
		if err != nil {
			return fmt.Errorf("writing env line %q: %v", v, err)
		}
	}
	return nil
}

func FindEnvFiles(root string) ([]string, error) {
	extension := ".env"
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == extension {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("searching for .env files: %v", err)
	}
	return paths, nil
}
