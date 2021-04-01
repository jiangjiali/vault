package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*NamespaceCreateCommand)(nil)
var _ cli.CommandAutocomplete = (*NamespaceCreateCommand)(nil)

type NamespaceCreateCommand struct {
	*BaseCommand
}

func (c *NamespaceCreateCommand) Synopsis() string {
	return "创建新的NameSpace"
}

func (c *NamespaceCreateCommand) Help() string {
	helpText := `
使用: vault namespace create [选项] PATH

  创建子命名空间。
  创建的名称空间将相对于VAULT_NAMESPACE环境变量
  或 -namespace CLI标志中提供的名称空间。

  创建子命名空间（例如ns1/）：

      $ vault namespace create ns1

  从父名称空间创建子名称空间（例如ns1/ns2/）：

      $ vault namespace create -namespace=ns1 ns2

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *NamespaceCreateCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)
}

func (c *NamespaceCreateCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFolders()
}

func (c *NamespaceCreateCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *NamespaceCreateCommand) Run(args []string) int {
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

	secret, err := client.Logical().Write("sys/namespaces/"+namespacePath, nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error creating namespace: %s", err))
		return 2
	}

	// Handle single field output
	if c.flagField != "" {
		return PrintRawField(c.UI, secret, c.flagField)
	}

	return OutputSecret(c.UI, secret)
}
