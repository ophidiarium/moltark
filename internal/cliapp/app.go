package cliapp

import (
	"io"

	"github.com/mitchellh/cli"
	"github.com/ophidiarium/moltark/internal/command"
)

const version = "0.1.0-dev"

type Config struct {
	Args       []string
	WorkingDir string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
}

func Run(cfg Config) int {
	meta := command.Meta{
		WorkingDir: cfg.WorkingDir,
		Stdin:      cfg.Stdin,
		Stdout:     cfg.Stdout,
		Stderr:     cfg.Stderr,
	}

	commands := map[string]cli.CommandFactory{
		"init": func() (cli.Command, error) {
			return &command.InitCommand{Meta: meta}, nil
		},
		"plan": func() (cli.Command, error) {
			return &command.PlanCommand{Meta: meta}, nil
		},
		"apply": func() (cli.Command, error) {
			return &command.ApplyCommand{Meta: meta}, nil
		},
		"show": func() (cli.Command, error) {
			return &command.ShowCommand{Meta: meta}, nil
		},
		"doctor": func() (cli.Command, error) {
			return &command.DoctorCommand{Meta: meta}, nil
		},
		"version": func() (cli.Command, error) {
			return &command.VersionCommand{Version: version}, nil
		},
	}

	app := cli.NewCLI("moltark", version)
	app.Args = cfg.Args
	app.Commands = commands
	app.HelpFunc = cli.FilteredHelpFunc([]string{
		"init",
		"plan",
		"apply",
		"show",
		"doctor",
		"version",
	}, cli.BasicHelpFunc("moltark"))
	app.HelpWriter = cfg.Stdout

	exitCode, err := app.Run()
	if err != nil {
		_, _ = io.WriteString(cfg.Stderr, err.Error()+"\n")
		return 1
	}

	return exitCode
}
