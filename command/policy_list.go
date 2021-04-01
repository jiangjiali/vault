package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PolicyListCommand)(nil)
var _ cli.CommandAutocomplete = (*PolicyListCommand)(nil)

type PolicyListCommand struct {
	*BaseCommand
}

func (c *PolicyListCommand) Synopsis() string {
	return "列出已安装的策略"
}

func (c *PolicyListCommand) Help() string {
	helpText := `
使用: vault policy list [选项]

  列出安装在安全库服务器上的策略的名称。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PolicyListCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *PolicyListCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *PolicyListCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PolicyListCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	switch {
	case len(args) > 0:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 0, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	policies, err := client.Sys().ListPolicies()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error listing policies: %s", err))
		return 2
	}

	switch Format(c.UI) {
	case "table":
		for _, p := range policies {
			c.UI.Output(p)
		}
		return 0
	default:
		return OutputData(c.UI, policies)
	}
}
