package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*OperatorCommand)(nil)

type OperatorCommand struct {
	*BaseCommand
}

func (c *OperatorCommand) Synopsis() string {
	return "执行操作员特定任务"
}

func (c *OperatorCommand) Help() string {
	helpText := `
使用: vault operator <子命令> [选项] [参数]

  此命令将与Vault交互的操作员的子命令分组。
  大多数用户不需要与这些命令交互。
  以下是操作员命令的几个示例：

  初始化新的安全库群集：

      $ vault operator init

  迫使安全库辞去集群中的领导职务：

      $ vault operator step-down

  旋转安全库的基础加密密钥：

      $ vault operator rotate

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *OperatorCommand) Run(args []string) int {
	return cli.RunResultHelp
}
