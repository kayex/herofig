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

func Write(w io.Writer, config map[string]string, separator string) error {
	if separator == "" {
		separator = "\n"
	}

	for key, value := range config {
		line := key + "=" + value + separator
		_, err := fmt.Fprint(w, line)
		if err != nil {
			return fmt.Errorf("failed writing env line %s: %v", line, err)
		}
	}
	return nil
}
