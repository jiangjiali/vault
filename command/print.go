package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PrintCommand)(nil)
var _ cli.CommandAutocomplete = (*PrintCommand)(nil)

type PrintCommand struct {
	*BaseCommand
}

func (c *PrintCommand) Synopsis() string {
	return "打印运行时配置"
}

func (c *PrintCommand) Help() string {
	helpText := `
使用: vault print <子命令>

	此命令将用于与服务器运行时值交互的子命令分组。

子命令:
	token    当前正在使用的令牌
`
	return strings.TrimSpace(helpText)
}

func (c *PrintCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *PrintCommand) AutocompleteFlags() complete.Flags {
	return nil
}

func (c *PrintCommand) Run(args []string) int {
	return cli.RunResultHelp
}
