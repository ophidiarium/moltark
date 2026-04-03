package command

import (
	"flag"

	"github.com/ophidiarium/moltark/internal/moltark"
)

type ApplyCommand struct {
	Meta
}

func (c *ApplyCommand) Run(args []string) int {
	fs := flag.NewFlagSet("moltark apply", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	autoApprove := fs.Bool("auto-approve", false, "skip interactive approval")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	service := c.service()
	plan, err := service.Plan(c.WorkingDir)
	if err != nil {
		c.errorf("Error: %s\n", err)
		return 1
	}

	c.printLine(moltark.RenderPlan(plan))
	if plan.HasConflicts() {
		c.errorf("Apply aborted due to conflicts.\n")
		return 1
	}

	if !plan.HasActionableChanges() {
		c.printLine("")
		c.printLine("No changes. Repository is already reconciled.")
		return 0
	}

	if !*autoApprove {
		c.printLine("")
		if !c.confirm() {
			c.errorf("Apply cancelled.\n")
			return 1
		}
	}

	result, err := service.Apply(c.WorkingDir, plan)
	if err != nil {
		c.errorf("Error: %s\n", err)
		return 1
	}

	c.printLine("")
	c.printLine(moltark.RenderApply(result))
	return 0
}

func (c *ApplyCommand) Help() string {
	return `Usage: moltark apply [options]

  Reconcile the repository by applying the current Moltark plan.

Options:

  -auto-approve    Skip interactive approval before writing files.`
}

func (c *ApplyCommand) Synopsis() string {
	return "Apply the current Moltark reconciliation plan"
}
