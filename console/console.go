package console

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"os"
)

var Success = color.New(color.FgGreen).SprintfFunc()
var Warning = color.New(color.FgYellow).SprintfFunc()
var Error = color.New(color.FgRed).SprintfFunc()
var ConfigKey = color.New(color.FgMagenta).SprintfFunc()
var ConfigKeyHighlighted = color.New(color.FgMagenta, color.BgBlack).SprintfFunc()
var ConfigValue = color.New().SprintfFunc()
var App = color.New(color.FgMagenta).SprintfFunc()
var FilePath = color.New(color.FgCyan).SprintfFunc()
var ID = color.New(color.FgGreen).SprintfFunc()

func Confirm(message, prompt string, def bool) bool {
	fmt.Printf("%s ", Warning(message))

	if def {
		fmt.Print(Warning(fmt.Sprintf("%s [Y/n] ", prompt)))
	} else {
		fmt.Print(Warning(fmt.Sprintf("%s [y/N] ", prompt)))
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	if text == "\n" {
		return def
	}
	return text == "y\n" || text == "Y\n"
}

func ConfirmOverwrite(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return Confirm(fmt.Sprintf("The file %s already exists.", filename), "Overwrite?", false)
	}
	return true
}

func Fatal(v ...interface{}) {
	fmt.Print(v...)
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	fmt.Println(v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
	os.Exit(1)
}
