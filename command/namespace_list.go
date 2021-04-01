package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*NamespaceListCommand)(nil)
var _ cli.CommandAutocomplete = (*NamespaceListCommand)(nil)

type NamespaceListCommand struct {
	*BaseCommand
}

func (c *NamespaceListCommand) Synopsis() string {
	return "列出子NameSpaces"
}

func (c *NamespaceListCommand) Help() string {
	helpText := `
使用: vault namespaces list [选项]

  列出启用的子命名空间。

  列出所有启用的子命名空间：

      $ vault namespaces list

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *NamespaceListCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *NamespaceListCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFolders()
}

func (c *NamespaceListCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *NamespaceListCommand) Run(args []string) int {
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

	secret, err := client.Logical().List("sys/namespaces")
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error listing namespaces: %s", err))
		return 2
	}

	_, ok := extractListData(secret)
	if Format(c.UI) != "table" {
		if secret == nil || secret.Data != nil || !ok {
			OutputData(c.UI, map[string]interface{}{})
			return 2
		}
	}

	if secret == nil {
		c.UI.Error(fmt.Sprintf("No namespaces found"))
		return 2
	}

	// There could be e.g. warnings
	if secret.Data == nil {
		return OutputSecret(c.UI, secret)
	}

	if secret.WrapInfo != nil && secret.WrapInfo.TTL != 0 {
		return OutputSecret(c.UI, secret)
	}

	if !ok {
		c.UI.Error(fmt.Sprintf("No entries found"))
		return 2
	}

	return OutputList(c.UI, secret)
}
