package command

type VersionCommand struct {
	Meta
	Version string
}

func (c *VersionCommand) Run(args []string) int {
	c.printLine(c.Version)
	return 0
}

func (c *VersionCommand) Help() string {
	return `Usage: moltark version

  Print the Moltark version.`
}

func (c *VersionCommand) Synopsis() string {
	return "Print the Moltark version"
}
