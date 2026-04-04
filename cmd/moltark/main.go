package main

import (
	"os"

	"github.com/ophidiarium/moltark/internal/cliapp"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	os.Exit(cliapp.Run(cliapp.Config{
		Args:       os.Args[1:],
		WorkingDir: wd,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}))
}
