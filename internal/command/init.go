package command

import (
	"flag"
	"fmt"

	"github.com/ophidiarium/moltark/internal/model"
	"github.com/ophidiarium/moltark/internal/module"
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

	result, err := module.InitRepository(c.WorkingDir)
	if err != nil {
		c.errorf("Error: %s\n", err)
		return 1
	}

	c.printLine(result)
	return 0
}

func (c *InitCommand) Help() string {
	return fmt.Sprintf(`Usage: moltark init

  Initialize Moltark in the current repository by creating a minimal
  %s when one does not already exist.

  This command does not reconcile pyproject.toml. Run "moltark plan"
  and "moltark apply" after initialization.`, model.ProjectSpecFileName)
}

func (c *InitCommand) Synopsis() string {
	return "Initialize Moltark in the current repository"
}
