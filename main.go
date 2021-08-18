package main

import (
	"flag"
	"fmt"
	"github.com/kayex/herofig/config"
	"github.com/kayex/herofig/console"
	"github.com/kayex/herofig/env"
	"github.com/kayex/herofig/heroku"
	"github.com/kayex/herofig/print"
	"log"
	"os"
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

	cfg := make(config.Config)
	for _, pair := range args {
		k, v, err := env.ParsePair(pair)
		if err != nil {
			l.Fatalf("failed parsing variables: %v", err)
		}
		cfg[k] = v
	}

	p.Printf("Setting ")
	i := 0
	for k := range cfg {
		if i > 0 {
			p.Print(", ")
		}
		p.ConfigKey(k)
		i++
	}
	p.Print(" on ")
	p.App(h.App())
	p.Printf("...\n")

	err := h.SetConfig(cfg)
	if err != nil {
		l.Fatalf("failed setting %s: %v", strings.Join(args, " "), err)
	}

	p.Success("Successfully set %d config %s\n", len(cfg), pluralize("variable", "", "s", len(cfg)))
}

func Pull(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	destination := ""
	if len(args) >= 1 {
		destination = args[0]

		if !console.ConfirmOverwrite(p, destination) {
			p.Error("Aborting\n")
			os.Exit(2)
		}
	}

	p.Print("Pulling config from ")
	p.App(h.App())
	p.Printf("...\n")

	cfg, err := h.Config()
	if err != nil {
		l.Fatalf("failed pulling config: %v", err)
	}

	if destination == "" {
		for k, v := range cfg {
			p.ConfigKey(k)
			p.Print("=")
			p.ConfigValue(v)
			p.Print("\n")
		}
		return
	}

	err = env.Save(destination, cfg)
	if err != nil {
		l.Fatalf("failed saving config to %s: %v", destination, err)
	}

	p.Success("Pulled %d configuration variables into ", len(cfg))
	p.LocalFile(destination)
	p.Newline()
}

func Push(l *log.Logger, p *print.Printer, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		p.Println("Usage: herofig push [env file]")
		os.Exit(1)
	}
	source := args[0]

	config, err := env.Open(source)
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

	config, err := env.Open(source)
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

	cfg, err := h.Config()
	if err != nil {
		l.Fatalf("failed getting config from application: %v", err)
	}

	cmp := substringSearch

	for k, v := range cfg {
		indices := cmp(k, query)
		if len(indices) > 0 {
		IterateRunes:
			// Iterate over individual runes to apply highlighting to characters matched by the search.
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
	localEnvFiles, err := env.Find(l, ".")
	if err != nil {
		l.Fatal(fmt.Errorf("failed searching for .env files: %v", err))
	}
	for _, envFile := range localEnvFiles {
		localCfg, err := env.Open(envFile)
		if err != nil {
			l.Fatal(err)
		}

		p.LocalFile(envFile)
		p.Space()
		hash := localCfg.Hash()
		p.Printf("%s %x", hash.Mnemonic(), hash)
		p.Newline()
	}
	if len(localEnvFiles) > 0 {
		p.Newline()
	}

	cfg, err := h.Config()
	if err != nil {
		l.Fatalf("failed getting config from application: %v", err)
	}
	p.App(h.App())
	p.Space()
	hash := cfg.Hash()
	p.Printf("%s %x", hash.Mnemonic(), hash)
	p.Newline()
}

func substringSearch(haystack, needle string) []int {
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

func pluralize(noun, singularSuffix, pluralSuffix string, count int) string {
	if count == 1 {
		return noun + singularSuffix
	}
	return noun + pluralSuffix
}
