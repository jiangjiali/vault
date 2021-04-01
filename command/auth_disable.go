package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*AuthDisableCommand)(nil)
var _ cli.CommandAutocomplete = (*AuthDisableCommand)(nil)

type AuthDisableCommand struct {
	*BaseCommand
}

func (c *AuthDisableCommand) Synopsis() string {
	return "禁用身份验证方法"
}

func (c *AuthDisableCommand) Help() string {
	helpText := `
使用: vault auth disable [选项] PATH

  禁用给定PATH上的现有身份验证方法。
  参数对应于装入的PATH，而不是TYPE。
  一旦认证方法被禁用，它的PATH就不能再用于身份验证。

  通过禁用的身份验证方法生成的所有访问令牌都将立即被吊销。
  此命令将阻止，直到所有令牌被吊销。

  在 userpass/ 处禁用认证方法：

      $ vault auth disable userpass/

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *AuthDisableCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *AuthDisableCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultAuths()
}

func (c *AuthDisableCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *AuthDisableCommand) Run(args []string) int {
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

	path := ensureTrailingSlash(sanitizePath(args[0]))

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	if err := client.Sys().DisableAuth(path); err != nil {
		c.UI.Error(fmt.Sprintf("Error disabling auth method at %s: %s", path, err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Disabled the auth method (if it existed) at: %s", path))
	return 0
}
