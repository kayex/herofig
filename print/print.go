package print

import "github.com/fatih/color"

func Warning(msg string, args ...interface{}) {
	color.Red(msg, args)
}

func Info(msg string, args ...interface{}) {
	color.White(msg, args)
}
