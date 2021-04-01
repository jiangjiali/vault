package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVListCommand)(nil)
var _ cli.CommandAutocomplete = (*KVListCommand)(nil)

type KVListCommand struct {
	*BaseCommand
}

func (c *KVListCommand) Synopsis() string {
	return "列出数据或机密"
}

func (c *KVListCommand) Help() string {
	helpText := `

使用: vault kv list [选项] PATH

  列出位于给定路径的安全库键值存储区中的数据。

  列出键值存储区“my-app”文件夹下的值：

      $ vault kv list secret/my-app/

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *KVListCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *KVListCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFolders()
}

func (c *KVListCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVListCommand) Run(args []string) int {
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
	mountPath, v2, err := isKVv2(path, client)
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	if v2 {
		path = addPrefixToVKVPath(path, mountPath, "metadata")
	}

	secret, err := client.Logical().List(path)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error listing %s: %s", path, err))
		return 2
	}

	_, ok := extractListData(secret)
	if Format(c.UI) != "table" {
		if secret == nil || secret.Data == nil || !ok {
			OutputData(c.UI, map[string]interface{}{})
			return 2
		}
	}

	if secret == nil || secret.Data == nil {
		c.UI.Error(fmt.Sprintf("No value found at %s", path))
		return 2
	}

	// If the secret is wrapped, return the wrapped response.
	if secret.WrapInfo != nil && secret.WrapInfo.TTL != 0 {
		return OutputSecret(c.UI, secret)
	}

	if !ok {
		c.UI.Error(fmt.Sprintf("No entries found at %s", path))
		return 2
	}

	return OutputList(c.UI, secret)
}
