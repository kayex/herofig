package main

import "flag"

type Variable struct {
	name string
	value string
}

func main() {
	// pull - Get config from Heroku and save to file
	// push append - Add config variables to Heroku without overwriting anything
	// push new - Add new config variables to Heroku without updating existing ones
	// push overwrite - Add config variables to Heroku, replacing existing ones
	// encrypt - Encrypt .env file
	// decrypt - Decrypt .env file

	command := flag.Arg(0)
	switch command {
	case "pull":
		pull()
	}

}

func pull() {
	config, err := Pull
}
