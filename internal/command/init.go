package command

import (
	"flag"

	"github.com/ophidiarium/moltark/internal/moltark"
)

type InitCommand struct {
	Meta
}

func (c *InitCommand) Run(args []string) int {
	fs := flag.NewFlagSet("moltark init", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	result, err := moltark.InitRepository(c.WorkingDir)
	if err != nil {
		c.errorf("Error: %s\n", err)
		return 1
	}

	c.printLine(result)
	return 0
}

func (c *InitCommand) Help() string {
	return `Usage: moltark init

  Initialize Moltark in the current repository by creating a minimal
  Starlark Moltarkfile when one does not already exist.

  This command does not reconcile pyproject.toml. Run "moltark plan"
  and "moltark apply" after initialization.`
}

func (c *InitCommand) Synopsis() string {
	return "Initialize Moltark in the current repository"
}
