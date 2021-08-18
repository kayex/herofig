package console

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"os"
)

type Color int
type PrintFunc func(format string, a ...interface{}) string

const (
	colorDefault Color = -1
	colorSuccess       = Color(color.FgGreen)
	colorWarning       = Color(color.FgYellow)
	colorError         = Color(color.FgRed)
	colorLocal         = Color(color.FgCyan)
	colorRemote        = Color(color.FgMagenta)
	colorID            = Color(color.FgGreen)
)

type Console struct {
	Output *Output
}

type Output struct {
	colors map[Color]*color.Color
}

func NewConsole() *Console {
	return &Console{&Output{make(map[Color]*color.Color)}}
}

func (c *Console) Confirm(message, prompt string, def bool) bool {
	fmt.Printf("%s ", c.Output.Warning()(message))

	if def {
		fmt.Print(c.Output.Warning()(fmt.Sprintf("%s [Y/n] ", prompt)))
	} else {
		fmt.Print(c.Output.Warning()(fmt.Sprintf("%s [y/N] ", prompt)))
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

func (o *Output) Success() PrintFunc {
	return o.color(colorSuccess).SprintfFunc()
}

func (o *Output) Warning() PrintFunc {
	return o.color(colorWarning).SprintfFunc()
}

func (o *Output) Error() PrintFunc {
	return o.color(colorError).SprintfFunc()
}

func (c *Console) Fatal(v ...interface{}) {
	fmt.Print(v...)
	os.Exit(1)
}

func (c *Console) Fatalf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
	os.Exit(1)
}

func (o *Output) App() PrintFunc {
	return o.color(colorRemote).SprintfFunc()
}

func (o *Output) FilePath() PrintFunc {
	return o.color(colorLocal).SprintfFunc()
}

func (o *Output) ConfigKey() PrintFunc {
	return o.color(colorRemote).SprintfFunc()
}

func (o *Output) ConfigKeyHighlighted() PrintFunc {
	col := *o.color(colorRemote)
	return col.Add(color.BgBlack).SprintfFunc()
}

func (o *Output) ConfigValue() PrintFunc {
	return o.color(colorDefault).SprintfFunc()
}

func (o *Output) ID() PrintFunc {
	return o.color(colorID).SprintfFunc()
}

func (o *Output) color(col Color) *color.Color {
	if existing, ok := o.colors[col]; ok {
		return existing
	}

	var newColor *color.Color
	if col == colorDefault {
		newColor = color.New()
	} else {
		newColor = color.New(color.Attribute(col))
	}

	o.colors[col] = newColor
	return newColor
}

func (o *Output) cprintf(col *color.Color, msg string, args []interface{}) {
	_, _ = col.Printf(msg, args...)
}
