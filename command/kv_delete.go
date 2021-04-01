package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVDeleteCommand)(nil)
var _ cli.CommandAutocomplete = (*KVDeleteCommand)(nil)

type KVDeleteCommand struct {
	*BaseCommand

	flagVersions []string
}

func (c *KVDeleteCommand) Synopsis() string {
	return "删除KV存储中的版本"
}

func (c *KVDeleteCommand) Help() string {
	helpText := `
使用: vault kv delete [选项] PATH

  删除键值存储区中提供的版本和路径的数据。版本控制的数据将不会完全删除，
  但会标记为已删除，并且不再在正常的get请求中返回。

  要删除键“foo”的最新版本：

      $ vault kv delete secret/foo

  要删除key foo的版本3：

      $ vault kv delete -versions=3 secret/foo

  要删除所有版本和元数据，请参见“vault kv metadata”子命令。

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *KVDeleteCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP)
	// Common Options
	f := set.NewFlagSet("命令选项")

	f.StringSliceVar(&StringSliceVar{
		Name:    "versions",
		Target:  &c.flagVersions,
		Default: nil,
		Usage:   `指定要删除的版本号。`,
	})

	return set
}

func (c *KVDeleteCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFiles()
}

func (c *KVDeleteCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVDeleteCommand) Run(args []string) int {
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

	var secret *api.Secret
	if v2 {
		secret, err = c.deleteV2(path, mountPath, client)
	} else {
		secret, err = client.Logical().Delete(path)
	}

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

func (c *KVDeleteCommand) deleteV2(path, mountPath string, client *api.Client) (*api.Secret, error) {
	var err error
	var secret *api.Secret
	switch {
	case len(c.flagVersions) > 0:
		path = addPrefixToVKVPath(path, mountPath, "delete")
		data := map[string]interface{}{
			"versions": kvParseVersionsFlags(c.flagVersions),
		}

		secret, err = client.Logical().Write(path, data)
	default:
		path = addPrefixToVKVPath(path, mountPath, "data")
		secret, err = client.Logical().Delete(path)
	}

	return secret, err
}
