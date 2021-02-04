package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kayex/configtool/encryption"
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

	var app = flag.String("app", "", "The name of the Heroku application.")
	var secret = flag.String("secret", "", "The file containing the encryption secret.")

	flag.Parse()
	command := flag.Arg(0)
	args := flag.Args()[1:]

	var p Platform
	if *app != "" {
		p = heroku.New(*app)
	}

	switch command {
	// Remote commands.
	case "get":
		get(l, p, args)
	case "set":
		set(l, p, args)
	case "pull":
		pull(l, p, args)
	case "push":
		push(l, p, args)
	// Local commands.
	case "keygen":
		keygen(l, args)
	case "encrypt":
		encrypt(l, *secret, args)
	case "decrypt":
		decrypt(l, *secret, args)
	default:
		fmt.Println("Usage: configtool get|set|pull|push|encrypt|decrypt|keygen")
		os.Exit(1)
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
		os.Exit(1)
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

	printRemoteSuccess(p.Name(), fmt.Sprintf("Pulled %d configuration variables to %s", len(config), dest))
}

func encrypt(l *log.Logger, keyFile string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: configtool encrypt [file]")
		os.Exit(1)
	}

	if keyFile == "" {
		fmt.Println("Usage: configtool --secret [keyfile] encrypt [file]")
		os.Exit(1)
	}

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("could not read keyfile %s: %v", keyFile, err)
	}

	filename := args[0]
	err = encryption.EncryptFile(key, filename)
	if err != nil {
		log.Fatalf("encryption failed: %v", err)
	}

	cs := encryption.Describe()
	printLocalSuccess(filename, fmt.Sprintf("Successfully encrypted %s using %s (key length %d).", filename, cs.Cipher, cs.KeyLength))
}

func decrypt(l *log.Logger, keyFile string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: configtool --secret [keyfile] decrypt [file]")
		os.Exit(1)
	}

	if keyFile == "" {
		fmt.Println("Usage: configtool --secret [keyfile] encrypt [file]")
		os.Exit(1)
	}

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("could not read keyfile %s: %v", keyFile, err)
	}

	filename := args[0]
	newFilename, err := encryption.DecryptFile(key, filename)
	if err != nil {
		log.Fatalf("encryption failed: %v", err)
	}

	printLocalSuccess(newFilename, fmt.Sprintf("Successfully decrypted %s.", filename))
}

func keygen(l *log.Logger, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: configtool keygen [output file]")
		os.Exit(1)
	}
	output := args[0]

	key := encryption.GenerateKey()
	err := ioutil.WriteFile(output, key, 0600)
	if err != nil {
		l.Fatalf("failed writing keyfile: %v", err)
	}

	printLocalSuccess(output, fmt.Sprintf("Successfully generated key with length %d.", encryption.Describe().KeyLength))
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

func printLocalSuccess(filename, message string) {
	fmt.Printf("OK [%s] %s\n", filename, message)
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
