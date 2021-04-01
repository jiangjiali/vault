package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PolicyCommand)(nil)

// PolicyCommand is a Command that holds the audit commands
type PolicyCommand struct {
	*BaseCommand
}

func (c *PolicyCommand) Synopsis() string {
	return "与策略交互"
}

func (c *PolicyCommand) Help() string {
	helpText := `
使用: vault policy <子命令> [选项] [参数]

  此命令将用于与策略交互的子命令分组。
  用户可以在安全库中写入、读取和列出策略。

  列出所有启用的策略：

      $ vault policy list

  从本地磁盘上的内容创建名为“my-policy”的策略：

      $ vault policy write my-policy ./my-policy.hcl

  删除名为“my-policy”的策略：

      $ vault policy delete my-policy

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *PolicyCommand) Run(args []string) int {
	return cli.RunResultHelp
}
