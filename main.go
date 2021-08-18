package main

import (
	"bufio"
	"crypto/sha1"
	"flag"
	"fmt"
	"github.com/kayex/herofig/env"
	"github.com/kayex/herofig/heroku"
	"github.com/kayex/herofig/print"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

func main() {
	l := log.New(os.Stderr, "", log.LstdFlags)
	p := print.NewPrinter()

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
		Get(l, p, h, args)
	case "set":
		Set(l, p, h, args)
	case "pull":
		Pull(l, p, h, args)
	case "push":
		Push(l, p, h, args)
	case "push:new":
		PushNew(l, p, h, args)
	case "search":
		Search(l, p, h, args)
	case "hash":
		Hash(l, p, h, args)
	default:
		p.Println("Usage: herofig get|set|pull|push|push:new|search|hash")
		os.Exit(1)
	}
}

func Get(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		p.Println("Usage: herofig get [key]")
		os.Exit(1)
	}
	key := args[0]

	v, err := h.ConfigValue(key)
	if err != nil {
		l.Fatalf("failed getting value for %s: %v", key, err)
	}
	p.Print(v)
}

func Set(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		p.Println("Usage: herofig set KEY=VALUE")
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

	p.Printf("Setting ")
	i := 0
	for k := range config {
		if i > 0 {
			p.Print(", ")
		}
		p.ConfigKey(k)
		i++
	}
	p.Print(" on ")
	p.App(h.App())
	p.Printf("...\n")

	err := h.SetConfig(config)
	if err != nil {
		l.Fatalf("failed setting %s: %v", strings.Join(args, " "), err)
	}

	p.Success("Successfully set %d config %s\n", len(config), pluralize("variable", "", "s", len(config)))
}

func Pull(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	destination := ""
	if len(args) >= 1 {
		destination = args[0]

		if !confirmOverwrite(p, destination) {
			p.Error("Aborting\n")
			os.Exit(2)
		}
	}

	p.Print("Pulling config from ")
	p.App(h.App())
	p.Printf("...\n")

	config, err := h.Config()
	if err != nil {
		l.Fatalf("failed pulling config: %v", err)
	}

	if destination == "" {
		for k, v := range config {
			p.ConfigKey(k)
			p.Print("=")
			p.ConfigValue(v)
			p.Print("\n")
		}
		return
	}

	err = writeEnvFile(destination, config)
	if err != nil {
		l.Fatalf("failed saving config to %s: %v", destination, err)
	}

	p.Success("Pulled %d configuration variables into ", len(config))
	p.LocalFile(destination)
	p.Newline()
}

func Push(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		p.Println("Usage: herofig push [env file]")
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

	p.Success(fmt.Sprintf("Successfully pushed %d configuration %s.", len(config), pluralize("variable", "", "s", len(config))))
}

func PushNew(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		p.Println("Usage: herofig push:new [env file]")
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

	p.Success(fmt.Sprintf("Successfully pushed %d new configuration %s.", len(config), pluralize("variable", "", "s", len(config))))
}

func Search(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		p.Println("Usage: herofig search [query]")
	}
	query := args[0]

	config, err := h.Config()
	if err != nil {
		l.Fatalf("failed getting config from application: %v", err)
	}

	for k, v := range config {
		indices := substringIndexes(k, query)
		if len(indices) > 0 {
		IterateRunes:
			// Iterate over runes to apply alternative output styling to characters matched by the search.
			for pos, r := range []rune(k) {
				rs := string(r)
				for _, i := range indices {
					if pos >= i && pos <= (i+utf8.RuneCountInString(query)-1) {
						p.ConfigKeyHighlighted(rs)
						continue IterateRunes
					}
				}
				p.ConfigKey(rs)
			}

			p.ConfigValue("=")
			p.ConfigValue(v)
			p.Newline()
		}
	}
}

func Hash(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	localEnvFiles := findEnvFiles(l, ".")
	for _, envFile := range localEnvFiles {
		localConfig, err := parseEnvFile(envFile)
		if err != nil {
			l.Fatal(err)
		}

		p.LocalFile(envFile)
		p.Space()
		p.Print(hash(localConfig))
		p.Newline()
	}
	if len(localEnvFiles) > 0 {
		p.Newline()
	}

	config, err := h.Config()
	if err != nil {
		l.Fatalf("failed getting config from application: %v", err)
	}
	p.App(h.App())
	p.Space()
	p.Print(hash(config))
	p.Newline()
}

func hash(config map[string]string) string {
	lines := make([]string, 0, len(config))
	for k, v := range config {
		lines = append(lines, env.Line(k, v))
	}
	sort.Strings(lines)

	h := sha1.New()
	for _, l := range lines {
		h.Write([]byte(l))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
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

	err = env.Write(f, config)
	if err != nil {
		return fmt.Errorf("failed writing to env file: %v", err)
	}
	return nil
}

func findEnvFiles(l *log.Logger, root string) []string {
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
		l.Fatal(fmt.Errorf("failed searching for .env files: %v", err))
	}
	return paths
}

func confirm(p *print.Printer, message, prompt string, def bool) bool {
	p.Warning(message)
	p.Print(" ")

	if def {
		p.Warning("%s [Y/n] ", prompt)
	} else {
		p.Warning("%s [y/N] ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	if text == "\n" {
		return def
	}
	return text == "y\n" || text == "Y\n"
}

func confirmOverwrite(p *print.Printer, filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return confirm(p, fmt.Sprintf("The file %s already exists.", filename), "Overwrite?", false)
	}
	return true
}

func pluralize(noun, singularSuffix, pluralSuffix string, count int) string {
	if count == 1 {
		return noun + singularSuffix
	}
	return noun + pluralSuffix
}

func substringIndexes(haystack, needle string) []int {
	var indices []int

	offset := 0
	for {
		i := strings.Index(haystack, needle)
		if i == -1 {
			break
		}

		indices = append(indices, offset+i)
		offset = offset + utf8.RuneCountInString(needle)
		haystack = string([]rune(haystack)[offset:])
	}

	return indices
}
