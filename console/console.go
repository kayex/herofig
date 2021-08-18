package console

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
)

type Color int

const (
	Default Color = -1
	Success       = Color(color.FgGreen)
	Warning       = Color(color.FgYellow)
	Error         = Color(color.FgRed)
	Local         = Color(color.FgCyan)
	Remote        = Color(color.FgMagenta)
)

type Console struct {
	colors map[Color]*color.Color
}

func NewConsole(l *log.Logger) *Console {
	return &Console{colors: make(map[Color]*color.Color)}
}

func (c *Console) Confirm(message, prompt string, def bool) bool {
	c.Warning(message)
	c.Print(" ")

	if def {
		c.Warning("%s [Y/n] ", prompt)
	} else {
		c.Warning("%s [y/N] ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	if text == "\n" {
		return def
	}
	return text == "y\n" || text == "Y\n"
}

func (c *Console) ConfirmOverwrite(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return c.Confirm(fmt.Sprintf("The file %s already exists.", filename), "Overwrite?", false)
	}
	return true
}

func (c *Console) Success(msg string, args ...interface{}) {
	c.cprintf(c.color(Success), msg, args)
}

func (c *Console) Warning(msg string, args ...interface{}) {
	c.cprintf(c.color(Warning), msg, args)
}

func (c *Console) Error(msg string, args ...interface{}) {
	c.cprintf(c.color(Error), msg, args)
}

func (c *Console) Fatal(err error) {
	log.Fatal(err)
}

func (c *Console) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

func (c *Console) Print(msg ...interface{}) {
	_, _ = c.color(Default).Print(msg...)
}

func (c *Console) Printf(msg string, args ...interface{}) {
	c.cprintf(c.color(Default), msg, args)
}

func (c *Console) Println(msg ...interface{}) {
	_, _ = c.color(Default).Println(msg...)
}

func (c *Console) PrintSpace() {
	_, _ = c.color(Default).Print(" ")
}

func (c *Console) PrintNewline() {
	_, _ = c.color(Default).Println()
}

func (c *Console) PrintApp(msg string, args ...interface{}) {
	c.cprintf(c.color(Remote), msg, args)
}

func (c *Console) PrintFilePath(msg string, args ...interface{}) {
	c.cprintf(c.color(Local), msg, args)
}

func (c *Console) PrintConfigKey(key string, args ...interface{}) {
	c.cprintf(c.color(Remote), key, args)
}

func (c *Console) PrintConfigKeyHighlighted(msg string, args ...interface{}) {
	col := *c.color(Remote)
	c.cprintf(col.Add(color.BgBlack), msg, args)
}

func (c *Console) PrintConfigValue(value string, args ...interface{}) {
	c.cprintf(c.color(Default), value, args)
}

func (c *Console) color(col Color) *color.Color {
	if existing, ok := c.colors[col]; ok {
		return existing
	}

	var newColor *color.Color
	if col == Default {
		newColor = color.New()
	} else {
		newColor = color.New(color.Attribute(col))
	}

	c.colors[col] = newColor
	return newColor
}

func (c *Console) cprintf(col *color.Color, msg string, args []interface{}) {
	_, _ = col.Printf(msg, args...)
}
