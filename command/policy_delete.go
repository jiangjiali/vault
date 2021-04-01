package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PolicyDeleteCommand)(nil)
var _ cli.CommandAutocomplete = (*PolicyDeleteCommand)(nil)

type PolicyDeleteCommand struct {
	*BaseCommand
}

func (c *PolicyDeleteCommand) Synopsis() string {
	return "按名称删除策略"
}

func (c *PolicyDeleteCommand) Help() string {
	helpText := `
使用: vault policy delete [选项] NAME

  删除安全库服务器中名为NAME的策略。
  删除策略后，与该策略关联的所有令牌将立即受到影响。

  删除名为“my-policy”的策略：

      $ vault policy delete my-policy

  请注意，不能删除“default”或“root”策略。这些都是内置的策略。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PolicyDeleteCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *PolicyDeleteCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultPolicies()
}

func (c *PolicyDeleteCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PolicyDeleteCommand) Run(args []string) int {
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

	name := strings.TrimSpace(strings.ToLower(args[0]))
	if err := client.Sys().DeletePolicy(name); err != nil {
		c.UI.Error(fmt.Sprintf("Error deleting %s: %s", name, err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Deleted policy: %s", name))
	return 0
}
