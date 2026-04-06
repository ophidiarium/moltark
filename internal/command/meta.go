package command

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/ophidiarium/moltark/internal/engine"
)

type Meta struct {
	WorkingDir string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
}

func (m Meta) service() engine.Service {
	return engine.NewService()
}

func (m Meta) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(m.Stdout, format, args...)
}

func (m Meta) errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(m.Stderr, format, args...)
}

func (m Meta) printLine(text string) {
	_, _ = io.WriteString(m.Stdout, text+"\n")
}

func (m Meta) confirm() bool {
	if m.Stdin == nil {
		return false
	}

	m.printLine("Do you want to perform these actions?")
	m.printLine("  Only 'yes' will be accepted to approve.")
	m.printLine("")
	_, _ = io.WriteString(m.Stdout, "Enter a value: ")

	reader := bufio.NewReader(m.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	return strings.TrimSpace(answer) == "yes"
}
