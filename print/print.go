package print

import (
	"github.com/fatih/color"
)

func Key(msg string, args ...interface{}) {
	c := color.New(color.FgMagenta)
	_, _ = c.Printf(msg, args...)
}

func Value(msg string, args ...interface{}) {
	c := color.New()
	_, _ = c.Printf(msg, args...)
}

func Success(msg string, args ...interface{}) {
	c := color.New(color.FgGreen)
	_, _ = c.Printf(msg, args...)
}

func Warning(msg string, args ...interface{}) {
	c := color.New(color.FgYellow)
	_, _ = c.Printf(msg, args...)
}

func Error(msg string, args ...interface{}) {
	c := color.New(color.FgRed)
	_, _ = c.Printf(msg, args...)
}
