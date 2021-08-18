package main

import (
	"flag"
	"fmt"
	"github.com/kayex/herofig/config"
	"github.com/kayex/herofig/console"
	"github.com/kayex/herofig/env"
	"github.com/kayex/herofig/heroku"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

func main() {
	l := log.New(os.Stderr, "", log.LstdFlags)
	c := console.NewConsole(l)

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
		Get(c, h, args)
	case "set":
		Set(c, h, args)
	case "pull":
		Pull(c, h, args)
	case "push":
		Push(c, h, args)
	case "push:new":
		PushNew(c, h, args)
	case "search":
		Search(c, h, args)
	case "hash":
		Hash(c, h, args)
	default:
		c.Println("Usage: herofig get|set|pull|push|push:new|search|hash")
		os.Exit(1)
	}
}

func Get(c *console.Console, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		c.Println("Usage: herofig get [key]")
		os.Exit(1)
	}
	key := args[0]

	v, err := h.ConfigValue(key)
	if err != nil {
		c.Fatalf("failed getting value for %s: %v", key, err)
	}
	c.Print(v)
}

func Set(c *console.Console, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		c.Println("Usage: herofig set KEY=VALUE")
		os.Exit(1)
	}

	cfg := make(config.Config)
	for _, pair := range args {
		k, v, err := env.ParsePair(pair)
		if err != nil {
			c.Fatalf("failed parsing variables: %v", err)
		}
		cfg[k] = v
	}

	c.Printf("Setting ")
	i := 0
	for k := range cfg {
		if i > 0 {
			c.Print(", ")
		}
		c.PrintConfigKey(k)
		i++
	}
	c.Print(" on ")
	c.PrintApp(h.App())
	c.Printf("...\n")

	err := h.SetConfig(cfg)
	if err != nil {
		c.Fatalf("failed setting %s: %v", strings.Join(args, " "), err)
	}

	c.Success("Successfully set %d configuration %s\n", len(cfg), pluralize("variable", "", "s", len(cfg)))
}

func Pull(c *console.Console, h *heroku.Heroku, args []string) {
	destination := ""
	if len(args) >= 1 {
		destination = args[0]

		if !c.ConfirmOverwrite(destination) {
			c.Error("Aborting\n")
			os.Exit(2)
		}
	}

	c.Print("Pulling configuration from ")
	c.PrintApp(h.App())
	c.Print("...\n")

	cfg, err := h.Config()
	if err != nil {
		c.Fatalf("failed pulling config: %v", err)
	}

	if destination == "" {
		for k, v := range cfg {
			c.PrintConfigKey(k)
			c.Print("=")
			c.PrintConfigValue(v)
			c.Print("\n")
		}
		return
	}

	err = env.Save(destination, cfg)
	if err != nil {
		c.Fatalf("failed saving config to %s: %v", destination, err)
	}

	c.Success("Pulled %d configuration variables into ", len(cfg))
	c.PrintFilePath(destination)
	c.PrintNewline()
}

func Push(c *console.Console, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		c.Println("Usage: herofig push [env file]")
		os.Exit(1)
	}
	source := args[0]

	cfg, err := env.Open(source)
	if err != nil {
		c.Fatal(err)
	}

	err = h.SetConfig(cfg)
	if err != nil {
		c.Fatalf("failed pushing config: %v", err)
	}

	c.Success(fmt.Sprintf("Successfully pushed %d configuration %s.", len(cfg), pluralize("variable", "", "s", len(cfg))))
}

func PushNew(c *console.Console, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		c.Println("Usage: herofig push:new [env file]")
		os.Exit(1)
	}
	source := args[0]

	existing, err := h.Config()
	if err != nil {
		c.Fatalf("failed getting existing config from application: %v", err)
	}

	cfg, err := env.Open(source)
	if err != nil {
		c.Fatal(err)
	}

	newConfig := make(map[string]string)

	for k, v := range cfg {
		if _, exists := existing[k]; !exists {
			newConfig[k] = v
		}
	}

	err = h.SetConfig(newConfig)
	if err != nil {
		c.Fatalf("failed pushing config to application: %v", err)
	}

	c.Success(fmt.Sprintf("Successfully pushed %d new configuration %s.", len(cfg), pluralize("variable", "", "s", len(cfg))))
}

func Search(c *console.Console, h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		c.Println("Usage: herofig search [query]")
	}
	query := args[0]

	cfg, err := h.Config()
	if err != nil {
		c.Fatalf("failed getting config from application: %v", err)
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
						c.PrintConfigKeyHighlighted(rs)
						continue IterateRunes
					}
				}
				c.PrintConfigKey(rs)
			}

			c.PrintConfigValue("=")
			c.PrintConfigValue(v)
			c.PrintNewline()
		}
	}
}

func Hash(c *console.Console, h *heroku.Heroku, args []string) {
	localEnvFiles, err := env.Find(".")
	if err != nil {
		c.Fatal(fmt.Errorf("failed searching for .env files: %v", err))
	}
	for _, envFile := range localEnvFiles {
		localCfg, err := env.Open(envFile)
		if err != nil {
			c.Fatal(err)
		}

		c.PrintFilePath(envFile)
		c.PrintSpace()
		hash := localCfg.Hash()
		c.PrintID(hash.Mnemonic(2))
		c.PrintSpace()
		c.Printf("%x", hash)
		c.PrintNewline()
	}
	if len(localEnvFiles) > 0 {
		c.PrintNewline()
	}

	cfg, err := h.Config()
	if err != nil {
		c.Fatalf("failed getting config from application: %v", err)
	}
	c.PrintApp(h.App())
	c.PrintSpace()
	hash := cfg.Hash()
	c.PrintID(hash.Mnemonic(2))
	c.PrintSpace()
	c.Printf("%x", hash)
	c.PrintNewline()
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
