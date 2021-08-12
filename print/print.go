package print

import (
	"github.com/fatih/color"
)

func p(c *color.Color, msg string, args []interface{}) {
	_, _ = c.Printf(msg, args...)
}

func Key(msg string, args ...interface{}) {
	p(color.New(color.FgMagenta), msg, args)
}

func Value(msg string, args ...interface{}) {
	p(color.New(), msg, args)
}

func Success(msg string, args ...interface{}) {
	p(color.New(color.FgGreen), msg, args)
}

func Warning(msg string, args ...interface{}) {
	p(color.New(color.FgYellow), msg, args)
}

func Error(msg string, args ...interface{}) {
	p(color.New(color.FgRed), msg, args)
}
