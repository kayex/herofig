package env

import (
	"strings"
)

func ParsePair(pair string) (string, string) {
	// TODO: Validation
	parts := strings.Split(pair, "=")
	return parts[0], parts[1]
}

func KeyValue(key, value string) string {
	return key + "=" + value
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
