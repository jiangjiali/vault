package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*NamespaceLookupCommand)(nil)
var _ cli.CommandAutocomplete = (*NamespaceLookupCommand)(nil)

type NamespaceLookupCommand struct {
	*BaseCommand
}

func (c *NamespaceLookupCommand) Synopsis() string {
	return "查找现有NameSpace"
}

func (c *NamespaceLookupCommand) Help() string {
	helpText := `
使用: vault namespace create [选项] PATH

  创建子命名空间。
  创建的名称空间将相对于VAULT_NAMESPACE环境变量
  或 -namespace CLI标志中提供的名称空间。

  获取有关本地身份验证令牌的命名空间的信息：

      $ vault namespace lookup

  获取有关特定子令牌（例如ns1/ns2/）的名称空间的信息：

      $ vault namespace lookup -namespace=ns1 ns2

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *NamespaceLookupCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *NamespaceLookupCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFolders()
}

func (c *NamespaceLookupCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *NamespaceLookupCommand) Run(args []string) int {
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

	namespacePath := strings.TrimSpace(args[0])

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	secret, err := client.Logical().Read("sys/namespaces/" + namespacePath)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error looking up namespace: %s", err))
		return 2
	}
	if secret == nil {
		c.UI.Error("Namespace not found")
		return 2
	}

	return OutputSecret(c.UI, secret)
}
