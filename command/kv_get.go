package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVGetCommand)(nil)
var _ cli.CommandAutocomplete = (*KVGetCommand)(nil)

type KVGetCommand struct {
	*BaseCommand

	flagVersion int
}

func (c *KVGetCommand) Synopsis() string {
	return "从KV存储中检索数据"
}

func (c *KVGetCommand) Help() string {
	helpText := `
使用: vault kv get [选项] KEY

  从安全库的键值存储中检索给定密钥名称的值。如果不存在具有该名称的键，
  则返回错误。如果存在具有该名称但没有数据的键，则不返回任何内容。

      $ vault kv get secret/foo

  要及时查看特定版本的给定密钥名称，请指定“-version”标志：

      $ vault kv get -version=1 secret/foo

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *KVGetCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)

	// Common Options
	f := set.NewFlagSet("命令选项")

	f.IntVar(&IntVar{
		Name:    "version",
		Target:  &c.flagVersion,
		Default: 0,
		Usage:   `如果通过，将返回版本号处的值。`,
	})

	return set
}

func (c *KVGetCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *KVGetCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVGetCommand) Run(args []string) int {
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

	var versionParam map[string]string

	if v2 {
		path = addPrefixToVKVPath(path, mountPath, "data")

		if c.flagVersion > 0 {
			versionParam = map[string]string{
				"version": fmt.Sprintf("%d", c.flagVersion),
			}
		}
	}

	secret, err := kvReadRequest(client, path, versionParam)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading %s: %s", path, err))
		if secret != nil {
			OutputSecret(c.UI, secret)
		}
		return 2
	}
	if secret == nil {
		c.UI.Error(fmt.Sprintf("No value found at %s", path))
		return 2
	}

	if c.flagField != "" {
		if v2 {
			// This is a v2, pass in the data field
			if data, ok := secret.Data["data"]; ok && data != nil {
				// If they requested a literal "data" see if they meant actual
				// value or the data block itself
				if c.flagField == "data" {
					if dataMap, ok := data.(map[string]interface{}); ok {
						if _, ok := dataMap["data"]; ok {
							return PrintRawField(c.UI, dataMap, c.flagField)
						}
					}
					return PrintRawField(c.UI, secret, c.flagField)
				}
				return PrintRawField(c.UI, data, c.flagField)
			} else {
				c.UI.Error(fmt.Sprintf("No data found at %s", path))
				return 2
			}
		} else {
			return PrintRawField(c.UI, secret, c.flagField)
		}
	}

	// If we have wrap info print the secret normally.
	if secret.WrapInfo != nil || c.flagFormat != "table" {
		return OutputSecret(c.UI, secret)
	}

	if len(secret.Warnings) > 0 {
		tf := TableFormatter{}
		tf.printWarnings(c.UI, secret)
	}

	if metadata, ok := secret.Data["metadata"]; ok && metadata != nil {
		c.UI.Info(getHeaderForMap("Metadata", metadata.(map[string]interface{})))
		OutputData(c.UI, metadata)
		c.UI.Info("")
	}

	data := secret.Data
	if v2 && data != nil {
		data = nil
		dataRaw := secret.Data["data"]
		if dataRaw != nil {
			data = dataRaw.(map[string]interface{})
		}
	}

	if data != nil {
		c.UI.Info(getHeaderForMap("Data", data))
		OutputData(c.UI, data)
	}

	return 0
}
