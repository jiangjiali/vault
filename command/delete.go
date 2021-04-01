package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*DeleteCommand)(nil)
var _ cli.CommandAutocomplete = (*DeleteCommand)(nil)

type DeleteCommand struct {
	*BaseCommand
}

func (c *DeleteCommand) Synopsis() string {
	return "删除机密和配置"
}

func (c *DeleteCommand) Help() string {
	helpText := `
使用: vault delete [选项] PATH

  从给定路径的服务器中删除机密和配置。
  “delete”的行为被委托给与给定路径相对应的后端。

  删除状态机密后端中的数据：

      $ vault delete secret/my-secret

  在传输后端卸载加密密钥：

      $ vault delete transit/keys/my-key

  删除 ops 角色：

      $ vault delete userpass/users/ops

  有关示例和路径的完整列表，请参阅与正在使用的秘密后端相对应的文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *DeleteCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *DeleteCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFiles()
}

func (c *DeleteCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *DeleteCommand) Run(args []string) int {
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

	secret, err := client.Logical().Delete(path)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error deleting %s: %s", path, err))
		if secret != nil {
			OutputSecret(c.UI, secret)
		}
		return 2
	}

	c.UI.Info(fmt.Sprintf("Success! Data deleted (if it existed) at: %s", path))
	return 0
}
