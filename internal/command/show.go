package command

import (
	"encoding/json"
	"flag"
)

type ShowCommand struct {
	Meta
}

func (c *ShowCommand) Run(args []string) int {
	fs := flag.NewFlagSet("moltark show", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	report, err := c.service().Show(c.WorkingDir)
	if err != nil {
		c.errorf("Error: %s\n", err)
		return 1
	}

	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		c.errorf("Error: %s\n", err)
		return 1
	}

	c.printf("%s\n", output)
	return 0
}

func (c *ShowCommand) Help() string {
	return `Usage: moltark show

  Print the Moltark engine pipeline artifacts as JSON, including the
  evaluated model, resolved capabilities, inspected repository state,
  planned changes, and candidate persisted state.`
}

func (c *ShowCommand) Synopsis() string {
	return "Show Moltark pipeline artifacts and state"
}
