package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/kayex/herofig/env"
	"github.com/kayex/herofig/heroku"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var errUserCancel = errors.New("user canceled")

type Platform interface {
	Name() string
	Get() (map[string]string, error)
	GetValue(key string) (string, error)
	Set(map[string]string) error
	SetValue(key, value string) error
}

func main() {
	// pull - Get config from Heroku and save to file
	// push - Add config variables to Heroku without overwriting anything
	// push:new - Add new config variables to Heroku without updating existing ones
	// push:overwrite - Add config variables to Heroku, replacing existing ones

	start := time.Now()
	l := log.New(os.Stderr, "", log.LstdFlags)

	// Accept explicit application name using -a and --app flags to stay consistent with the Heroku CLI.
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
	// TODO: Add push:overwrite
	default:
		fmt.Println("Usage: herofig get|set|pull|push|push:new")
		os.Exit(1)
	}

	duration := time.Now().Sub(start)
	fmt.Printf("Finished in %v\n", round(duration, 2))
}

func get(l *log.Logger, p Platform, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig get [key]")
		os.Exit(1)
	}
	key := args[0]

	v, err := p.GetValue(key)
	if err != nil {
		l.Fatalf("failed getting value for %s: %v", key, err)
	}
	fmt.Println(v)
}

func set(l *log.Logger, p Platform, args []string) {
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

	fmt.Printf("Setting config on %s...\n", p.Name())

	err := p.Set(config)
	if err != nil {
		l.Fatalf("failed setting %s: %v", strings.Join(args, " "), err)
	}

	printSuccess(p.Name(), strings.Join(args, " "))
}

func pushNew(l *log.Logger, p Platform, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig push:new [source file]")
		os.Exit(1)
	}
	source := args[0]

	existing, err := p.Get()
	if err != nil {
		l.Fatalf("failed getting existing configuration from Heroku: %v", err)
	}

	config, err := readEnvSource(source)
	if err != nil {
		l.Fatal(err)
	}

	newConfig := make(map[string]string)

	for k, v := range config {
		if _, exists := existing[k]; !exists {
			newConfig[k] = v
		}
	}

	err = p.Set(newConfig)
	if err != nil {
		l.Fatalf("failed pushing configuration to Heroku: %v", err)
	}

	printSuccess(p.Name(), fmt.Sprintf("Successfully pushed %d new configuration %s.", len(config), pluralize("variable", len(config))))
}

func push(l *log.Logger, p Platform, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig push [source file]")
		os.Exit(1)
	}
	source := args[0]

	config, err := readEnvSource(source)
	if err != nil {
		l.Fatal(err)
	}

	err = p.Set(config)
	if err != nil {
		l.Fatalf("failed pushing config: %v", err)
	}

	printSuccess(p.Name(), fmt.Sprintf("Successfully pushed %d configuration %s.", len(config), pluralize("variables", len(config))))
}

func pull(l *log.Logger, p Platform, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig pull [target file]")
		os.Exit(1)
	}
	dest := args[0]

	if !confirmOverwrite(dest) {
		fmt.Println("Aborting")
		os.Exit(2)
	}

	fmt.Printf("Pulling config from %s...\n", p.Name())

	config, err := p.Get()
	if err != nil {
		l.Fatalf("failed pulling config: %v", err)
	}

	if dest == "" {
		fmt.Printf("%s\n", config)
		return
	}

	err = export(l, config, dest)
	if err != nil {
		l.Fatalf("failed saving config to %s: %v", dest, err)
	}

	printSuccess(p.Name(), fmt.Sprintf("Pulled %d configuration variables into %s", len(config), dest))
}

func readEnvSource(filename string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read source file %v: %v", filename, err)
	}

	config, err := env.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing source file: %v", err)
	}

	return config, nil
}

func confirm(message, prompt string, def bool) bool {
	fmt.Println(message)

	if def {
		fmt.Printf("%s [Y/n] ", prompt)
	} else {
		fmt.Printf("%s [y/N] ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	if text == "\n" {
		return def
	}
	return text == "y\n" || text == "Y\n"
}

func confirmOverwrite(dest string) bool {
	if _, err := os.Stat(dest); err == nil {
		return confirm(fmt.Sprintf("The file %s already exists. Pass --overwrite to force overwrite.", dest), "Overwrite?", false)
	}
	return true
}

func export(l *log.Logger, config map[string]string, dest string) error {
	return ioutil.WriteFile(dest, env.FromConfig(config, "\n"), 0644)
}

func printSuccess(app, message string) {
	fmt.Printf("OK [%s] %s\n", app, message)
}

func round(d time.Duration, digits int) time.Duration {
	var divs = []time.Duration{time.Duration(1), time.Duration(10), time.Duration(100), time.Duration(1000)}
	switch {
	case d > time.Second:
		d = d.Round(time.Second / divs[digits])
	case d > time.Millisecond:
		d = d.Round(time.Millisecond / divs[digits])
	case d > time.Microsecond:
		d = d.Round(time.Microsecond / divs[digits])
	}
	return d
}

func pluralize(word string, count int) string {
	dict := map[string]struct {
		One  string
		Many string
	}{
		"variable": {
			One:  "variable",
			Many: "variables",
		},
	}

	if f, ok := dict[word]; ok {
		if count == 1 {
			return f.One
		}
		return f.Many
	}

	return word
}
