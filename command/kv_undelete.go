package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVUndeleteCommand)(nil)
var _ cli.CommandAutocomplete = (*KVUndeleteCommand)(nil)

type KVUndeleteCommand struct {
	*BaseCommand

	flagVersions []string
}

func (c *KVUndeleteCommand) Synopsis() string {
	return "恢复KV存储中的版本"
}

func (c *KVUndeleteCommand) Help() string {
	helpText := `
使用: vault kv undelete [选项] KEY

  在键值存储区中恢复所提供版本和路径的数据。
  这将恢复数据，允许在get请求时返回数据。

  要撤消删除密钥“foo”的版本3：
  
      $ vault kv undelete -versions=3 secret/foo

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *KVUndeleteCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	// Common Options
	f := set.NewFlagSet("命令选项")

	f.StringSliceVar(&StringSliceVar{
		Name:    "versions",
		Target:  &c.flagVersions,
		Default: nil,
		Usage:   `指定要撤消删除的版本号。`,
	})

	return set
}

func (c *KVUndeleteCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *KVUndeleteCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVUndeleteCommand) Run(args []string) int {
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

	if len(c.flagVersions) == 0 {
		c.UI.Error("No versions provided, use the \"-versions\" flag to specify the version to undelete.")
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
		c.UI.Error("Undelete not supported on KV Version 1")
		return 1
	}

	path = addPrefixToVKVPath(path, mountPath, "undelete")
	data := map[string]interface{}{
		"versions": kvParseVersionsFlags(c.flagVersions),
	}

	secret, err := client.Logical().Write(path, data)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error writing data to %s: %s", path, err))
		if secret != nil {
			OutputSecret(c.UI, secret)
		}
		return 2
	}
	if secret == nil {
		// Don't output anything unless using the "table" format
		if Format(c.UI) == "table" {
			c.UI.Info(fmt.Sprintf("Success! Data written to: %s", path))
		}
		return 0
	}

	return OutputSecret(c.UI, secret)
}
