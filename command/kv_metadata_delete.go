package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVMetadataDeleteCommand)(nil)
var _ cli.CommandAutocomplete = (*KVMetadataDeleteCommand)(nil)

type KVMetadataDeleteCommand struct {
	*BaseCommand
}

func (c *KVMetadataDeleteCommand) Synopsis() string {
	return "删除KV存储区中密钥的所有版本和元数据"
}

func (c *KVMetadataDeleteCommand) Help() string {
	helpText := `
使用: vault kv metadata delete [选项] PATH

  删除所提供密钥的所有版本和元数据。

      $ vault kv metadata delete secret/foo

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *KVMetadataDeleteCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *KVMetadataDeleteCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFiles()
}

func (c *KVMetadataDeleteCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVMetadataDeleteCommand) Run(args []string) int {
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
	mountPath, v2, err := isKVv2(path, client)
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}
	if !v2 {
		c.UI.Error("Metadata not supported on KV Version 1")
		return 1
	}

	path = addPrefixToVKVPath(path, mountPath, "metadata")
	if secret, err := client.Logical().Delete(path); err != nil {
		c.UI.Error(fmt.Sprintf("Error deleting %s: %s", path, err))
		if secret != nil {
			OutputSecret(c.UI, secret)
		}
		return 2
	}

	c.UI.Info(fmt.Sprintf("Success! Data deleted (if it existed) at: %s", path))
	return 0
}
