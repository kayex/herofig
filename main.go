package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kayex/herofig/env"
	"github.com/kayex/herofig/heroku"
	"github.com/kayex/herofig/print"
	"log"
	"os"
	"strings"
)

func main() {
	l := log.New(os.Stderr, "", log.LstdFlags)

	// Accept explicit application name using -a and --app flags to be consistent with the Heroku CLI.
	var a = flag.String("a", "", "The Heroku application name.")
	var app = flag.String("app", "", "The Heroku application name.")
	flag.Parse()
	command := flag.Arg(0)
	args := flag.Args()[1:]
	if *a == "" {
		a = app
	}

	h := heroku.NewHeroku(*a)

	switch command {
	case "get":
		get(l, h, args)
	case "set":
		set(l, h, args)
	case "pull":
		pull(l, h, args)
	case "push":
		push(l, h, args)
	case "push:new":
		pushNew(l, h, args)
	default:
		fmt.Println("Usage: herofig get|set|pull|push|push:new")
		os.Exit(1)
	}
}

func get(l *log.Logger, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig get [key]")
		os.Exit(1)
	}
	key := args[0]

	v, err := h.ConfigValue(key)
	if err != nil {
		l.Fatalf("failed getting value for %s: %v", key, err)
	}
	fmt.Println(v)
}

func set(l *log.Logger, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig set KEY=VALUE")
		os.Exit(1)
	}

	config := make(map[string]string)
	for _, pair := range args {
		k, v, err := env.ParsePair(pair)
		if err != nil {
			l.Fatalf("failed parsing variables: %v", err)
		}
		config[k] = v
	}

	fmt.Printf("Setting config on %s...\n", h.App())

	err := h.SetConfig(config)
	if err != nil {
		l.Fatalf("failed setting %s: %v", strings.Join(args, " "), err)
	}

	success(h.App(), strings.Join(args, " "))
}

func pull(l *log.Logger, h *heroku.Heroku, args []string) {
	destination := ""
	if len(args) >= 1 {
		destination = args[0]

		if !confirmOverwrite(destination) {
			print.Error("Aborting\n")
			os.Exit(2)
		}
	}

	fmt.Printf("Pulling config from %s...\n", h.App())

	config, err := h.Config()
	if err != nil {
		l.Fatalf("failed pulling config: %v", err)
	}

	if destination == "" {
		for k, v := range config {
			print.Key(k)
			fmt.Print("=")
			print.Value(v)
			fmt.Print("\n")
		}
		return
	}

	err = writeEnvFile(destination, config)
	if err != nil {
		l.Fatalf("failed saving config to %s: %v", destination, err)
	}

	success(h.App(), fmt.Sprintf("Pulled %d configuration variables into %s", len(config), destination))
}

func push(l *log.Logger, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig push [env file]")
		os.Exit(1)
	}
	source := args[0]

	config, err := parseEnvFile(source)
	if err != nil {
		l.Fatal(err)
	}

	err = h.SetConfig(config)
	if err != nil {
		l.Fatalf("failed pushing config: %v", err)
	}

	success(h.App(), fmt.Sprintf("Successfully pushed %d configuration %s.", len(config), pluralize("variable", "", "s", len(config))))
}

func pushNew(l *log.Logger, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig push:new [env file]")
		os.Exit(1)
	}
	source := args[0]

	existing, err := h.Config()
	if err != nil {
		l.Fatalf("failed getting existing config from application: %v", err)
	}

	config, err := parseEnvFile(source)
	if err != nil {
		l.Fatal(err)
	}

	newConfig := make(map[string]string)

	for k, v := range config {
		if _, exists := existing[k]; !exists {
			newConfig[k] = v
		}
	}

	err = h.SetConfig(newConfig)
	if err != nil {
		l.Fatalf("failed pushing config to application: %v", err)
	}

	success(h.App(), fmt.Sprintf("Successfully pushed %d new configuration %s.", len(config), pluralize("variable", "", "s", len(config))))
}

func parseEnvFile(filename string) (map[string]string, error) {
	data, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read env file %v: %v", filename, err)
	}
	defer data.Close()

	config, err := env.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing env file: %v", err)
	}

	return config, nil
}

func writeEnvFile(filename string, config map[string]string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed open env file for writing: %v", err)
	}

	err = env.Write(f, config, "\n")
	if err != nil {
		return fmt.Errorf("failed writing to env file: %v", err)
	}
	return nil
}

func confirm(message, prompt string, def bool) bool {
	print.Warning(message)
	fmt.Print(" ")

	if def {
		print.Warning("%s [Y/n] ", prompt)
	} else {
		print.Warning("%s [y/N] ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	if text == "\n" {
		return def
	}
	return text == "y\n" || text == "Y\n"
}

func confirmOverwrite(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return confirm(fmt.Sprintf("The file %s already exists.", filename), "Overwrite?", false)
	}
	return true
}

func success(app, message string) {
	print.Success("OK [%s] %s\n", app, message)
}

func pluralize(word, singularSuffix, pluralSuffix string, count int) string {
	if count == 1 {
		return word + singularSuffix
	}
	return word + pluralSuffix
}
