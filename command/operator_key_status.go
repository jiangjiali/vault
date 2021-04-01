package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*OperatorKeyStatusCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorKeyStatusCommand)(nil)

type OperatorKeyStatusCommand struct {
	*BaseCommand
}

func (c *OperatorKeyStatusCommand) Synopsis() string {
	return "提供有关活动加密密钥的信息"
}

func (c *OperatorKeyStatusCommand) Help() string {
	helpText := `
使用: vault operator key-status [选项]

  提供有关活动加密密钥的信息。具体地说，当前密钥项和密钥安装时间。
` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorKeyStatusCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *OperatorKeyStatusCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *OperatorKeyStatusCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorKeyStatusCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	if len(args) > 0 {
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 0, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	status, err := client.Sys().KeyStatus()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading key status: %s", err))
		return 2
	}

	switch Format(c.UI) {
	case "table":
		c.UI.Output(printKeyStatus(status))
		return 0
	default:
		return OutputData(c.UI, status)
	}
}
