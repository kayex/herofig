package print

import (
	"github.com/fatih/color"
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

type Printer struct {
	colors map[Color]*color.Color
}

func NewPrinter() *Printer {
	return &Printer{colors: make(map[Color]*color.Color)}
}

func (p *Printer) color(c Color) *color.Color {
	if existing, ok := p.colors[c]; ok {
		return existing
	}

	var newColor *color.Color
	if c == Default {
		newColor = color.New()
	} else {
		newColor = color.New(color.Attribute(c))
	}

	p.colors[c] = newColor
	return newColor
}

func (p *Printer) Print(msg ...interface{}) {
	_, _ = p.color(Default).Print(msg...)
}

func (p *Printer) Printf(msg string, args ...interface{}) {
	_, _ = p.color(Default).Printf(msg, args...)
}

func (p *Printer) Println(msg ...interface{}) {
	_, _ = p.color(Default).Println(msg...)
}

func (p *Printer) Newline() {
	_, _ = p.color(Default).Println()
}

func (p *Printer) OK(id string) {
	p.Success("OK [")
	p.Remote(id)
	p.Success("] ")
}

func (p *Printer) Remote(msg string, args ...interface{}) {
	p.cprintf(p.color(Remote), msg, args)
}

func (p *Printer) Local(msg string, args ...interface{}) {
	p.cprintf(p.color(Local), msg, args)
}

func (p *Printer) ConfigKey(key string, args ...interface{}) {
	p.cprintf(p.color(Remote), key, args)
}

func (p *Printer) ConfigKeyHighlighted(msg string, args ...interface{}) {
	c := *p.color(Remote)
	p.cprintf(c.Add(color.BgBlack), msg, args)
}

func (p *Printer) ConfigValue(value string, args ...interface{}) {
	p.cprintf(p.color(Default), value, args)
}

func (p *Printer) Success(msg string, args ...interface{}) {
	p.cprintf(p.color(Success), msg, args)
}

func (p *Printer) Warning(msg string, args ...interface{}) {
	p.cprintf(p.color(Warning), msg, args)
}

func (p *Printer) Error(msg string, args ...interface{}) {
	p.cprintf(p.color(Error), msg, args)
}

func (p *Printer) cprintf(c *color.Color, msg string, args []interface{}) {
	_, _ = c.Printf(msg, args...)
}
