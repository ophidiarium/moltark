package command

import (
	"flag"

	"github.com/ophidiarium/moltark/internal/moltark"
)

type DoctorCommand struct {
	Meta
}

func (c *DoctorCommand) Run(args []string) int {
	fs := flag.NewFlagSet("moltark doctor", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	report, err := c.service().Doctor(c.WorkingDir)
	if err != nil {
		c.errorf("Error: %s\n", err)
		return 1
	}

	c.printLine(moltark.RenderDoctor(report))
	if report.HasIssues {
		return 1
	}

	return 0
}

func (c *DoctorCommand) Help() string {
	return `Usage: moltark doctor

  Validate the Moltark configuration, state, and managed repository
  surfaces, and report drift or conflicts.`
}

func (c *DoctorCommand) Synopsis() string {
	return "Validate repository state and managed ownership"
}
