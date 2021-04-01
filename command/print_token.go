package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PrintTokenCommand)(nil)
var _ cli.CommandAutocomplete = (*PrintTokenCommand)(nil)

type PrintTokenCommand struct {
	*BaseCommand
}

func (c *PrintTokenCommand) Synopsis() string {
	return "打印当前使用的安全库令牌"
}

func (c *PrintTokenCommand) Help() string {
	helpText := `
使用: vault print token

  在考虑配置的令牌助手和环境后，打印将用于命令的安全库令牌的值。

      $ vault print token

`
	return strings.TrimSpace(helpText)
}

func (c *PrintTokenCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *PrintTokenCommand) AutocompleteFlags() complete.Flags {
	return nil
}

func (c *PrintTokenCommand) Run(args []string) int {
	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	c.UI.Output(client.Token())
	return 0
}
