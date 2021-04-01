package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PolicyReadCommand)(nil)
var _ cli.CommandAutocomplete = (*PolicyReadCommand)(nil)

type PolicyReadCommand struct {
	*BaseCommand
}

func (c *PolicyReadCommand) Synopsis() string {
	return "打印策略的内容"
}

func (c *PolicyReadCommand) Help() string {
	helpText := `
使用: vault policy read [选项] [NAME]

  打印名为NAME的安全库策略的内容和元数据。如果策略不存在，则返回错误。

  阅读名为“my-policy”的政策：

      $ vault policy read my-policy

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PolicyReadCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *PolicyReadCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultPolicies()
}

func (c *PolicyReadCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PolicyReadCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	switch {
	case len(args) < 1:
		c.UI.Error(fmt.Sprintf("Not enough arguments (expected 1, got %d)", len(args)))
		return 1
	case len(args) > 1:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	name := strings.ToLower(strings.TrimSpace(args[0]))
	rules, err := client.Sys().GetPolicy(name)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading policy named %s: %s", name, err))
		return 2
	}
	if rules == "" {
		c.UI.Error(fmt.Sprintf("No policy named: %s", name))
		return 2
	}

	switch Format(c.UI) {
	case "table":
		c.UI.Output(strings.TrimSpace(rules))
		return 0
	default:
		resp := map[string]string{
			"policy": rules,
		}
		return OutputData(c.UI, &resp)
	}
}
