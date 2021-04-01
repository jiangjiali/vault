package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*FinanceCommand)(nil)

type FinanceCommand struct {
	*BaseCommand
}

func (c *FinanceCommand) Synopsis() string {
	return "金融计算公式"
}

func (c *FinanceCommand) Help() string {
	helpText := `
使用: vault finance <子命令> [选项]

  此命令包含用于与安全库的键值存储区交互的子命令。
  下面是一些简单的示例，更详细的示例可以在子命令或文档中找到。

  债券：

      $ vault finance bonds [选项]

  现金流量：

      $ vault finance cashflow [选项]

  折旧：

      $ vault finance depreciation [选项]

  利率：

      $ vault finance rates [选项]

  时间价值：

      $ vault finance tvm [选项]

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *FinanceCommand) Run(args []string) int {
	return cli.RunResultHelp
}
