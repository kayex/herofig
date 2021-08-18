package console

import (
	"bufio"
	"fmt"
	"github.com/kayex/herofig/print"
	"os"
)

func Confirm(p *print.Printer, message, prompt string, def bool) bool {
	p.Warning(message)
	p.Print(" ")

	if def {
		p.Warning("%s [Y/n] ", prompt)
	} else {
		p.Warning("%s [y/N] ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	if text == "\n" {
		return def
	}
	return text == "y\n" || text == "Y\n"
}

func ConfirmOverwrite(p *print.Printer, filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return Confirm(p, fmt.Sprintf("The file %s already exists.", filename), "Overwrite?", false)
	}
	return true
}
