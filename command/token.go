package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*TokenCommand)(nil)

type TokenCommand struct {
	*BaseCommand
}

func (c *TokenCommand) Synopsis() string {
	return "与令牌交互"
}

func (c *TokenCommand) Help() string {
	helpText := `
使用: vault token <子命令> [选项] [参数]

  此命令将用于与令牌交互的子命令分组。用户可以创建、查找、续订和吊销令牌。

  创建新令牌：

      $ vault token create

  吊销令牌：

      $ vault token revoke 96ddf4bc-d217-f3ba-f9bd-017055595017

  续订令牌：

      $ vault token renew 96ddf4bc-d217-f3ba-f9bd-017055595017

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *TokenCommand) Run(args []string) int {
	return cli.RunResultHelp
}
