package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVMetadataPutCommand)(nil)
var _ cli.CommandAutocomplete = (*KVMetadataPutCommand)(nil)

type KVMetadataPutCommand struct {
	*BaseCommand

	flagMaxVersions int
	flagCASRequired bool
	testStdin       io.Reader // for tests
}

func (c *KVMetadataPutCommand) Synopsis() string {
	return "设置或更新KV存储中的密钥元数据"
}

func (c *KVMetadataPutCommand) Help() string {
	helpText := `
使用: vault metadata kv put [选项] KEY

  此命令可用于在键值存储区中创建空密钥，或更新指定密钥的密钥配置。
  
  在键值存储区中创建一个没有数据的键：

      $ vault kv metadata put secret/foo

  在键上设置最大版本设置：

      $ vault kv metadata put -max-versions=5 secret/foo

  需要检查并设置此密钥：

      $ vault kv metadata put -cas-required secret/foo

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *KVMetadataPutCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	// Common Options
	f := set.NewFlagSet("命令选项")

	f.IntVar(&IntVar{
		Name:    "max-versions",
		Target:  &c.flagMaxVersions,
		Default: 0,
		Usage:   `要保留的版本数。如果未设置，则使用后端配置的最高版本。`,
	})

	f.BoolVar(&BoolVar{
		Name:    "cas-required",
		Target:  &c.flagCASRequired,
		Default: false,
		Usage:   `如果为true，则键将要求在所有写请求上设置cas参数。如果为false，将使用后端的配置。`,
	})

	return set
}

func (c *KVMetadataPutCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *KVMetadataPutCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVMetadataPutCommand) Run(args []string) int {
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
	data := map[string]interface{}{
		"max_versions": c.flagMaxVersions,
		"cas_required": c.flagCASRequired,
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
