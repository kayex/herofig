package env

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func ParsePair(pair string) (string, string, error) {
	parts := strings.Split(pair, "=")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid variable format: %v", pair)
	}

	return parts[0], parts[1], nil
}

func KeyValue(key, value string) string {
	return key + "=" + value
}

func Parse(env io.Reader) (map[string]string, error) {
	config := make(map[string]string)

	scanner := bufio.NewScanner(env)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		k, v, err := ParsePair(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("error processing line: %v", err)
		}
		config[k] = v
	}
	return config, nil
}

func FromConfig(config map[string]string, separator string) []byte {
	if separator == "" {
		separator = "\n"
	}

	var env string
	for key, value := range config {
		env += key + "=" + value + separator
	}
	return []byte(env)
}
