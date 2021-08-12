package print

import "github.com/fatih/color"

func Success(msg string, args ...interface{}) {
	color.Green(msg, args...)
}

func Warning(msg string, args ...interface{}) {
	color.Yellow(msg, args...)
}

func Error(msg string, args ...interface{}) {
	color.Red(msg, args...)
}

func Info(msg string, args ...interface{}) {
	color.White(msg, args...)
}
