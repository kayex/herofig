package env

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func ParsePair(pair string) (string, string, error) {
	delimiter := strings.Index(pair, "=")
	if delimiter < 1 {
		return "", "", fmt.Errorf("invalid env variable format: %v", pair)
	}
	pr := []rune(pair)

	key := string(pr[:delimiter])
	value := string(pr[delimiter+1:])

	return key, value, nil
}

func KeyValue(key, value string) string {
	return key + "=" + value
}

func Line(key, value string) string {
	return KeyValue(key, value) + "\n"
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

func Write(w io.Writer, config map[string]string) error {
	for key, value := range config {
		line := Line(key, value)
		_, err := fmt.Fprint(w, line)
		if err != nil {
			return fmt.Errorf("failed writing env line %s: %v", line, err)
		}
	}
	return nil
}
