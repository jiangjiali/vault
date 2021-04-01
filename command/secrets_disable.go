package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*SecretsDisableCommand)(nil)
var _ cli.CommandAutocomplete = (*SecretsDisableCommand)(nil)

type SecretsDisableCommand struct {
	*BaseCommand
}

func (c *SecretsDisableCommand) Synopsis() string {
	return "关闭机密引擎"
}

func (c *SecretsDisableCommand) Help() string {
	helpText := `
使用: vault secrets disable [选项] PATH

  禁用给定PATH上的机密引擎。参数对应于引擎的启用PATH，而不是TYPE！
  此引擎创建的所有机密都将被吊销，其安全库数据也将被删除。

  在 kv/ 时禁用机密引擎：

      $ vault secrets disable kv/

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *SecretsDisableCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *SecretsDisableCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultMounts()
}

func (c *SecretsDisableCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *SecretsDisableCommand) Run(args []string) int {
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

	path := ensureTrailingSlash(sanitizePath(args[0]))

	if err := client.Sys().Unmount(path); err != nil {
		c.UI.Error(fmt.Sprintf("Error disabling secrets engine at %s: %s", path, err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Disabled the secrets engine (if it existed) at: %s", path))
	return 0
}
