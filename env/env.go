package env

import (
	"bufio"
	"fmt"
	"github.com/kayex/herofig/config"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func KeyValue(key, value string) string {
	return key + "=" + value
}

func Line(key, value string) string {
	return KeyValue(key, value) + "\n"
}

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

func Parse(env io.Reader) (config.Config, error) {
	cfg := make(map[string]string)

	scanner := bufio.NewScanner(env)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		k, v, err := ParsePair(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("error processing line: %v", err)
		}
		cfg[k] = v
	}
	return cfg, nil
}

func Write(w io.Writer, cfg config.Config) error {
	for key, value := range cfg {
		line := Line(key, value)
		_, err := fmt.Fprint(w, line)
		if err != nil {
			return fmt.Errorf("failed writing env line %s: %v", line, err)
		}
	}
	return nil
}

func Open(filename string) (config.Config, error) {
	data, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read env file %v: %v", filename, err)
	}
	defer data.Close()

	cfg, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing env file: %v", err)
	}

	return cfg, nil
}

func Save(filename string, cfg config.Config) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed open env file for writing: %v", err)
	}

	err = Write(f, cfg)
	if err != nil {
		return fmt.Errorf("failed writing to env file: %v", err)
	}
	return nil
}

func Find(l *log.Logger, root string) ([]string, error) {
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
		return nil, fmt.Errorf("failed searching for .env files: %v", err)
	}
	return paths, nil
}
