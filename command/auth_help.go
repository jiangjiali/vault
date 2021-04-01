package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*AuthHelpCommand)(nil)
var _ cli.CommandAutocomplete = (*AuthHelpCommand)(nil)

type AuthHelpCommand struct {
	*BaseCommand

	Handlers map[string]LoginHandler
}

func (c *AuthHelpCommand) Synopsis() string {
	return "打印身份验证方法的用法"
}

func (c *AuthHelpCommand) Help() string {
	helpText := `
使用: vault auth help [选项] TYPE | PATH

  打印身份验证方法的用法和帮助。

    - 如果给定TYPE，此命令将打印该类型的认证方法的默认帮助。

    - 如果给定PATH，此命令将打印在该路径上启用的认证方法的帮助输出。
      此路径必须已存在。

  获取 userpass 身份验证方法的用法说明：

      $ vault auth help userpass

  在 my-method/ 上启用的认证方法的打印用法：

      $ vault auth help my-method/

  每个认证方法生成自己的帮助输出。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *AuthHelpCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *AuthHelpCommand) AutocompleteArgs() complete.Predictor {
	handlers := make([]string, 0, len(c.Handlers))
	for k := range c.Handlers {
		handlers = append(handlers, k)
	}
	return complete.PredictSet(handlers...)
}

func (c *AuthHelpCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *AuthHelpCommand) Run(args []string) int {
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

	// Start with the assumption that we have an auth type, not a path.
	authType := strings.TrimSpace(args[0])

	authHandler, ok := c.Handlers[authType]
	if !ok {
		// There was no auth type by that name, see if it's a mount
		auths, err := client.Sys().ListAuth()
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error listing auth methods: %s", err))
			return 2
		}

		authPath := ensureTrailingSlash(sanitizePath(args[0]))
		auth, ok := auths[authPath]
		if !ok {
			c.UI.Warn(fmt.Sprintf(
				"No auth method available on the server at %q", authPath))
			return 1
		}

		authHandler, ok = c.Handlers[auth.Type]
		if !ok {
			c.UI.Warn(wrapAtLength(fmt.Sprintf(
				"No method-specific CLI handler available for auth method %q",
				authType)))
			return 2
		}
	}

	c.UI.Output(authHandler.Help())
	return 0
}
