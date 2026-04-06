package command

import (
	"flag"
	"fmt"

	"github.com/ophidiarium/moltark/internal/engine"
	"github.com/ophidiarium/moltark/internal/model"
)

type PlanCommand struct {
	Meta
}

func (c *PlanCommand) Run(args []string) int {
	fs := flag.NewFlagSet("moltark plan", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	detailedExitCode := fs.Bool("detailed-exitcode", false, "return 2 when the plan contains changes")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	plan, err := c.service().Plan(c.WorkingDir)
	if err != nil {
		c.errorf("Error: %s\n", err)
		return 1
	}

	c.printLine(engine.RenderPlan(plan))

	if plan.HasConflicts() {
		return 1
	}

	if *detailedExitCode && plan.HasActionableChanges() {
		return 2
	}

	return 0
}

func (c *PlanCommand) Help() string {
	return fmt.Sprintf(`Usage: moltark plan [options]

  Compare the %s desired state with the current repository and
  show the planned file updates without writing them.

Options:

  -detailed-exitcode    Return 2 when changes are planned.`, model.ProjectSpecFileName)
}

func (c *PlanCommand) Synopsis() string {
	return fmt.Sprintf("Show repository changes required by %s", model.ProjectSpecFileName)
}
