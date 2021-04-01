package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*SecretsCommand)(nil)

type SecretsCommand struct {
	*BaseCommand
}

func (c *SecretsCommand) Synopsis() string {
	return "与机密引擎互动"
}

func (c *SecretsCommand) Help() string {
	helpText := `
使用: vault secrets <子命令> [选项] [参数]

  此命令将子命令分组，以便与安全库的机密引擎进行交互。
  每个秘密引擎的行为都不同。
  有关详细信息，请参阅文档。

  列出所有启用的机密引擎：

      $ vault secrets list

  启用新的机密引擎：

      $ vault secrets enable database

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *SecretsCommand) Run(args []string) int {
	return cli.RunResultHelp
}
