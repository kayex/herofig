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
)

func main() {
	// pull - Get config from Heroku and save to file
	// push append - Add config variables to Heroku without overwriting anything
	// push new - Add new config variables to Heroku without updating existing ones
	// push overwrite - Add config variables to Heroku, replacing existing ones
	// encrypt - Encrypt .env file
	// decrypt - Decrypt .env file

	l := log.New(os.Stderr, "", log.LstdFlags)
	flag.Parse()
	command := flag.Arg(0)
	switch command {
	case "get":
		get(l)
	case "set":
		set(l)
	case "pull":
		pull(l)
	}

}

func get(l *log.Logger) {
	app := flag.Arg(1)
	key := flag.Arg(2)

	v, err := heroku.GetValue(app, key)
	if err != nil {
		l.Fatalf("failed getting value for %s: %v", key, err)
	}
	fmt.Println(v)
}

func set(l *log.Logger) {
	args := flag.Args()[1:]
	app := args[0]
	pairs := args[1:]
	config := make(map[string]string)

	for _, p := range pairs {
		k, v := env.ParsePair(p)
		config[k] = v
	}

	fmt.Printf("Setting config on %s...\n", app)

	err := heroku.Set(app, config)
	if err != nil {
		l.Fatalf("failed setting %s: %v", strings.Join(pairs, " "), err)
	}

	printRemoteSuccess(app, strings.Join(pairs, " "))
}

func pull(l *log.Logger) {
	app := flag.Arg(1)
	dest := flag.Arg(2)

	if !confirmOverwrite(dest) {
		fmt.Println("Aborting")
		os.Exit(2)
	}

	fmt.Printf("Pulling config from %s...\n", app)

	config, err := heroku.Pull(app)
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

	printRemoteSuccess(app, fmt.Sprintf("Pulled %d configuration values to %s", len(config), dest))
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
	fmt.Printf("OK [%s] %s", app, message)
}
