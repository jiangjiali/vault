package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*NamespaceCommand)(nil)

type NamespaceCommand struct {
	*BaseCommand
}

func (c *NamespaceCommand) Synopsis() string {
	return "与NameSpace交互"
}

func (c *NamespaceCommand) Help() string {
	helpText := `
使用: vault namespace <子命令> [选项] [参数]

  此命令将用于与服务器命名空间交互的子命令分组。
  这些子命令在当前登录令牌所属的命名空间上下文中操作。

  列表启用的子命名空间：

      $ vault namespace list

  查找现有命名空间：

      $ vault namespace lookup

  创建新命名空间：

      $ vault namespace create

  删除现有命名空间：

      $ vault namespace delete

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *NamespaceCommand) Run(args []string) int {
	return cli.RunResultHelp
}
