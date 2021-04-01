package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PathHelpCommand)(nil)
var _ cli.CommandAutocomplete = (*PathHelpCommand)(nil)

var pathHelpVaultSealedMessage = strings.TrimSpace(`
错误：安全库已密封.

“path-help”命令要求打开保险库，以便知道机密引擎的安装点。
`)

type PathHelpCommand struct {
	*BaseCommand
}

func (c *PathHelpCommand) Synopsis() string {
	return "检索路径的API帮助"
}

func (c *PathHelpCommand) Help() string {
	helpText := `
使用: vault path-help [选项] PATH

  检索路径的API帮助。Vault中的所有端点都以降价格式提供内置帮助。
  这包括系统路径、秘密引擎和身份验证方法。

  获取装载在 database/:

      $ vault path-help database/

  响应对象将返回其他路径以检索帮助：

      $ vault path-help database/roles/

  每个机密引擎产生不同的帮助输出。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PathHelpCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *PathHelpCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictAnything // TODO: programatic way to invoke help
}

func (c *PathHelpCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PathHelpCommand) Run(args []string) int {
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

	path := sanitizePath(args[0])

	help, err := client.Help(path)
	if err != nil {
		if strings.Contains(err.Error(), "Vault is sealed") {
			c.UI.Error(pathHelpVaultSealedMessage)
		} else {
			c.UI.Error(fmt.Sprintf("Error retrieving help: %s", err))
		}
		return 2
	}

	c.UI.Output(help.Help)
	return 0
}
