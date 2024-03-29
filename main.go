package main

import (
	"flag"
	"fmt"
	"github.com/kayex/herofig/config"
	"github.com/kayex/herofig/console"
	"github.com/kayex/herofig/env"
	"github.com/kayex/herofig/heroku"
	"strings"
	"unicode/utf8"
)

func main() {
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
		Get(h, args)
	case "set":
		Set(h, args)
	case "pull":
		Pull(h, args)
	case "push":
		Push(h, args)
	case "push:new":
		PushNew(h, args)
	case "search":
		Search(h, args)
	case "hash":
		Hash(h, args)
	default:
		console.Fatalln("Usage: herofig get|set|pull|push|push:new|search|hash")
	}
}

func Get(h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		console.Fatalln("Usage: herofig get [key]")
	}
	key := args[0]

	v, err := h.ConfigValue(key)
	if err != nil {
		console.Fatalf("failed getting value for %s: %v", key, err)
	}
	fmt.Print(v)
}

func Set(h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		console.Fatalln("Usage: herofig set KEY=VALUE")
	}

	cfg := make(config.Config)
	for _, pair := range args {
		k, v, err := env.ParsePair(pair)
		if err != nil {
			console.Fatalf("failed parsing variables: %v", err)
		}
		cfg[k] = v
	}

	var keys []string
	for k := range cfg {
		keys = append(keys, console.ConfigKey(k))
	}

	fmt.Printf("Setting %s on %s...\n", strings.Join(keys, ", "), console.App(h.App()))

	err := h.SetConfig(cfg)
	if err != nil {
		console.Fatalf("failed setting %s: %v", strings.Join(args, " "), err)
	}

	fmt.Println(console.Success("Successfully set %d configuration %s", len(cfg), pluralize("variable", "", "s", len(cfg))))
}

func Pull(h *heroku.Heroku, args []string) {
	destination := ""
	if len(args) >= 1 {
		destination = args[0]
		if !console.ConfirmOverwrite(destination) {
			console.Fatalln(console.Error("Aborting"))
		}
	}

	fmt.Printf("Pulling configuration from %s...\n", console.App(h.App()))

	cfg, err := h.Config()
	if err != nil {
		console.Fatalf("failed pulling config: %v", err)
	}

	if destination == "" {
		for k, v := range cfg {
			fmt.Printf("%s=%s\n", console.ConfigKey(k), console.ConfigValue(v))
		}
		return
	}

	err = env.Save(destination, cfg)
	if err != nil {
		console.Fatalf("failed saving config to %s: %v", destination, err)
	}

	fmt.Println(console.Success(fmt.Sprintf("Pulled %d configuration variables into %s", len(cfg), console.FilePath(destination))))
}

func Push(h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		console.Fatalln("Usage: herofig push [env file]")
	}
	source := args[0]

	cfg, err := env.Open(source)
	if err != nil {
		console.Fatal(err)
	}

	err = h.SetConfig(cfg)
	if err != nil {
		console.Fatalf("failed pushing config: %v", err)
	}

	fmt.Println(console.Success("Successfully pushed %d configuration %s.", len(cfg), pluralize("variable", "", "s", len(cfg))))
}

func PushNew(h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		console.Fatalln("Usage: herofig push:new [env file]")
	}
	source := args[0]

	existing, err := h.Config()
	if err != nil {
		console.Fatalf("failed getting existing config from application: %v", err)
	}

	cfg, err := env.Open(source)
	if err != nil {
		console.Fatal(err)
	}

	newConfig := make(map[string]string)

	for k, v := range cfg {
		if _, exists := existing[k]; !exists {
			newConfig[k] = v
		}
	}

	if len(newConfig) == 0 {
		fmt.Println(console.Warning("No new configuration variables."))
		return
	}

	err = h.SetConfig(newConfig)
	if err != nil {
		console.Fatalf("failed pushing config to application: %v", err)
	}

	fmt.Println(console.Success("Successfully pushed %d new configuration %s.", len(newConfig), pluralize("variable", "", "s", len(newConfig))))
}

func Search(h *heroku.Heroku, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig search [query]")
	}
	query := args[0]

	cfg, err := h.Config()
	if err != nil {
		console.Fatalf("failed getting config from application: %v", err)
	}

	strategy := substringSearch

	for k, v := range cfg {
		indices := strategy(k, query)
		if len(indices) > 0 {
		IterateRunes:
			// Iterate over individual runes to apply highlighting to characters matched by the search.
			for pos, r := range []rune(k) {
				rs := string(r)
				for _, i := range indices {
					if pos >= i && pos < i+utf8.RuneCountInString(query) {
						fmt.Print(console.ConfigKeyHighlighted(rs))
						continue IterateRunes
					}
				}
				fmt.Print(console.ConfigKey(rs))
			}
			fmt.Printf("=%s\n", console.ConfigValue(v))
		}
	}
}

func Hash(h *heroku.Heroku, args []string) {
	localEnvFiles, err := env.Find(".")
	if err != nil {
		console.Fatal(fmt.Errorf("failed searching for .env files: %v", err))
	}
	for _, envFile := range localEnvFiles {
		localCfg, err := env.Open(envFile)
		if err != nil {
			console.Fatal(err)
		}

		hash := localCfg.Hash()
		fmt.Printf("%s %s %x\n", console.FilePath(envFile), console.ID(hash.Mnemonic(2)), hash)
	}
	if len(localEnvFiles) > 0 {
		fmt.Println()
	}

	cfg, err := h.Config()
	if err != nil {
		console.Fatalf("failed getting config from application: %v", err)
	}
	hash := cfg.Hash()
	fmt.Printf("%s %s %x\n", console.App(h.App()), console.ID(hash.Mnemonic(2)), hash)
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
		offset += utf8.RuneCountInString(needle)
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
