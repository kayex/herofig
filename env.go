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

func KeyValue(key, value string) string {
	return key + "=" + value
}

func ParsePair(pair string) (key string, value string, err error) {
	delimiter := strings.Index(pair, "=")
	if delimiter < 1 {
		return "", "", fmt.Errorf("invalid env variable format %q", pair)
	}
	pr := []rune(pair)

	key = string(pr[:delimiter])
	value = string(pr[delimiter+1:])
	return key, value, nil
}

func Parse(env io.Reader) (Config, error) {
	cfg := make(map[string]string)

	scanner := bufio.NewScanner(env)
	scanner.Split(bufio.ScanLines)

	line := 1
	for scanner.Scan() {
		k, v, err := ParsePair(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("processing line %d: %v", line, err)
		}
		cfg[k] = v
		line++
	}
	return cfg, nil
}

func Open(filename string) (Config, error) {
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

	for key, value := range cfg {
		line := KeyValue(key, value)
		_, err := fmt.Fprintln(f, line)
		if err != nil {
			return fmt.Errorf("writing env line %q: %v", line, err)
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
