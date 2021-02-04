package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kayex/configtool/env"
	"github.com/kayex/configtool/heroku"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type Platform interface {
	Name() string
	Get() (map[string]string, error)
	GetValue(key string) (string, error)
	Set(map[string]string) error
	SetValue(key, value string) error
}

func main() {
	// pull - Get config from Heroku and save to file
	// push append - Add config variables to Heroku without overwriting anything
	// push new - Add new config variables to Heroku without updating existing ones
	// push overwrite - Add config variables to Heroku, replacing existing ones
	// encrypt - Encrypt .env file
	// decrypt - Decrypt .env file

	start := time.Now()

	l := log.New(os.Stderr, "", log.LstdFlags)

	flag.Parse()
	command := flag.Arg(0)
	app := flag.Arg(1)
	args := flag.Args()[2:]
	p := heroku.New(app)

	switch command {
	case "get":
		get(l, p, args)
	case "set":
		set(l, p, args)
	case "pull":
		pull(l, p, args)
	case "push":
		push(l, p, args)
	default:
		fmt.Println("Usage: configtool get|set|pull|push")
	}

	duration := time.Now().Sub(start)
	fmt.Printf("Finished in %v\n", round(duration, 2))
}

func get(l *log.Logger, p Platform, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: configtool get [app] [key]")
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
		fmt.Println("Usage: configtool set [app] KEY=VALUE")
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

	printRemoteSuccess(p.Name(), strings.Join(args, " "))
}

func push(l *log.Logger, p Platform, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: configtool push [app] [source file]")
	}
	source := args[0]

	data, err := ioutil.ReadFile(source)
	if err != nil {
		l.Fatalf("Could not read source file %v: %v", source, err)
	}

	config, err := env.Parse(data)
	if err != nil {
		l.Fatalf("error reading source file: %v", err)
	}

	err = p.Set(config)
	if err != nil {
		l.Fatalf("failed pushing config: %v", err)
	}

	printRemoteSuccess(p.Name(), fmt.Sprintf("Successfully pushed %d configuration variables.", len(config)))
}

func pull(l *log.Logger, p Platform, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: configtool pull [app] [target file]")
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

	printRemoteSuccess(p.Name(), fmt.Sprintf("Pulled %d configuration variables to %s", len(config), dest))
}

func confirmOverwrite(dest string) bool {
	if _, err := os.Stat(dest); err == nil {
		fmt.Printf("The file %s already exists. Pass --overwrite to force overwrite.\n", dest)
		fmt.Printf("Overwrite? [y/N] ")

		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		return text == "y\n" || text == "Y\n"
	}
	return true
}

func export(l *log.Logger, config map[string]string, dest string) error {
	return ioutil.WriteFile(dest, env.FromConfig(config, "\n"), 0644)
}

func printRemoteSuccess(app, message string) {
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
