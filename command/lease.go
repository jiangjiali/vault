package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*LeaseCommand)(nil)

type LeaseCommand struct {
	*BaseCommand
}

func (c *LeaseCommand) Synopsis() string {
	return "与租约交互"
}

func (c *LeaseCommand) Help() string {
	helpText := `
使用: vault lease <子命令> [选项] [参数]

  此命令将用于与租约交互的子命令分组。用户可以撤销或续订租约。

  续租：

      $ vault lease renew database/creds/readonly/2f6a614c...

  撤销租约：

      $ vault lease revoke database/creds/readonly/2f6a614c...
`

	return strings.TrimSpace(helpText)
}

func (c *LeaseCommand) Run(args []string) int {
	return cli.RunResultHelp
}
