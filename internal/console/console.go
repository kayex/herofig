package console

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fatih/color"
)

var Success = c(color.FgGreen)
var Warning = c(color.FgYellow)
var Error = c(color.FgRed)
var ConfigKey = c(color.FgMagenta)
var ConfigKeyHighlighted = c(color.FgMagenta, color.BgBlack)
var ConfigValue = c()
var App = c(color.FgMagenta)
var FilePath = c(color.FgCyan)
var ID = c(color.FgGreen)

func c(a ...color.Attribute) func(format string, a ...interface{}) string {
	return color.New(a...).SprintfFunc()
}

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
		return Confirm(fmt.Sprintf("%s already exists.", filename), "Overwrite?", false)
	}
	return true
}

func Fatal(v ...any) {
	fmt.Print(v...)
	os.Exit(1)
}

func Fatalln(v ...any) {
	fmt.Println(v...)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	fmt.Printf(format, v...)
	os.Exit(1)
}
