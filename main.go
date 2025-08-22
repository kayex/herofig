package main

import (
	"flag"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kayex/herofig/internal/console"
)

func main() {
	usageMessage := "Usage: herofig [-a app] get|set|pull|push|push:new|search|hash"
	// Accept explicit application name using -a and --app flags to be consistent with the Heroku CLI.
	var a = flag.String("a", "", "The Heroku application name.")
	var app = flag.String("app", "", "The Heroku application name.")
	flag.Parse()
	command := flag.Arg(0)

	if command == "" {
		console.Fatalln(usageMessage)
	}

	args := flag.Args()[1:]

	if *a == "" {
		a = app
	}

	h := NewHeroku(*a)

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
		console.Fatalln(usageMessage)
	}
}

func Get(h *Heroku, args []string) {
	if len(args) < 1 {
		console.Fatalln("Usage: herofig get [key]")
	}
	key := args[0]

	v, err := h.ConfigValue(key)
	if err != nil {
		console.Fatalf("getting value: %v", err)
	}
	fmt.Print(v)
}

func Set(h *Heroku, args []string) {
	if len(args) < 1 {
		console.Fatalln("Usage: herofig set KEY=VALUE")
	}

	cfg := make(Config)
	for _, v := range args {
		k, v, err := ParseVar(v)
		if err != nil {
			console.Fatalf("parsing variables: %v", err)
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
		console.Fatalln(err.Error())
	}

	fmt.Println(console.Success("Successfully set %d configuration %s", len(cfg), pluralize("variable", "", "s", len(cfg))))
}

func Pull(h *Heroku, args []string) {
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
		console.Fatalf("pulling config: %v", err)
	}
	ordered := cfg.Ordered()

	if destination == "" {
		for _, v := range ordered {
			fmt.Printf("%s=%s\n", console.ConfigKey(v.Key), console.ConfigValue(v.Value))
		}
		return
	}

	err = Save(destination, cfg)
	if err != nil {
		console.Fatalf("saving config to %s: %v", destination, err)
	}

	fmt.Println(console.Success(fmt.Sprintf("Pulled %d configuration variables into %s", len(cfg), console.FilePath(destination))))
}

func Push(h *Heroku, args []string) {
	if len(args) < 1 {
		console.Fatalln("Usage: herofig push [env file]")
	}
	source := args[0]

	cfg, err := Load(source)
	if err != nil {
		console.Fatalln(err)
	}

	err = h.SetConfig(cfg)
	if err != nil {
		console.Fatalf("pushing config: %v", err)
	}

	fmt.Println(console.Success("Successfully pushed %d configuration %s.", len(cfg), pluralize("variable", "", "s", len(cfg))))
}

func PushNew(h *Heroku, args []string) {
	if len(args) < 1 {
		console.Fatalln("Usage: herofig push:new [env file]")
	}
	source := args[0]

	existing, err := h.Config()
	if err != nil {
		console.Fatalf("getting existing config from application: %v", err)
	}

	cfg, err := Load(source)
	if err != nil {
		console.Fatalln(err)
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
		console.Fatalf("pushing config to application: %v", err)
	}

	fmt.Println(console.Success("Successfully pushed %d new configuration %s.", len(newConfig), pluralize("variable", "", "s", len(newConfig))))
}

func Search(h *Heroku, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: herofig search [query]")
	}
	query := args[0]

	cfg, err := h.Config()
	if err != nil {
		console.Fatalf("getting config from application: %v", err)
	}

	for _, v := range cfg.Ordered() {
		indices := substringSearch(v.Key, query)
		if len(indices) > 0 {
		IterateRunes:
			// Iterate over individual runes to apply highlighting to characters matched by the search.
			for pos, r := range []rune(v.Key) {
				rs := string(r)
				for _, i := range indices {
					if pos >= i && pos < i+utf8.RuneCountInString(query) {
						fmt.Print(console.ConfigKeyHighlighted(rs))
						continue IterateRunes
					}
				}
				fmt.Print(console.ConfigKey(rs))
			}
			fmt.Printf("=%s\n", console.ConfigValue(v.Value))
		}
	}
}

func Hash(h *Heroku, args []string) {
	localEnvFiles, err := FindEnvFiles(".")
	if err != nil {
		console.Fatalf("searching for .env files: %v", err)
	}
	for _, envFile := range localEnvFiles {
		localCfg, err := Load(envFile)
		if err != nil {
			console.Fatalln(err)
		}

		hash := localCfg.Hash()
		fmt.Printf("%s %s %x\n", console.FilePath(envFile), console.ID(hash.Mnemonic(2)), hash)
	}
	if len(localEnvFiles) > 0 {
		fmt.Println()
	}

	cfg, err := h.Config()
	if err != nil {
		console.Fatalf("getting config from application: %v", err)
	}
	hash := cfg.Hash()
	fmt.Printf("%s %s %x\n", console.App(h.App()), console.ID(hash.Mnemonic(2)), hash)
}

func substringSearch(haystack, needle string) []int {
	haystack = strings.ToLower(haystack)
	needle = strings.ToLower(needle)
	var indices []int

	for {
		i := strings.Index(haystack, needle)
		if i == -1 {
			break
		}

		indices = append(indices, i+len(needle)*len(indices))
		haystack = haystack[i+len(needle):]
	}

	return indices
}

func pluralize(noun, singularSuffix, pluralSuffix string, count int) string {
	if count == 1 {
		return noun + singularSuffix
	}
	return noun + pluralSuffix
}
