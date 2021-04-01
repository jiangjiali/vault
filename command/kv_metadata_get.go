package command

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVMetadataGetCommand)(nil)
var _ cli.CommandAutocomplete = (*KVMetadataGetCommand)(nil)

type KVMetadataGetCommand struct {
	*BaseCommand
}

func (c *KVMetadataGetCommand) Synopsis() string {
	return "从KV存储检索密钥元数据"
}

func (c *KVMetadataGetCommand) Help() string {
	helpText := `
使用: vault kv metadata get [选项] KEY

  从安全库的键值存储（位于给定的密钥名称）中检索元数据。
  如果不存在具有该名称的键，则返回错误。

      $ vault kv metadata get secret/foo

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *KVMetadataGetCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	return set
}

func (c *KVMetadataGetCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *KVMetadataGetCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVMetadataGetCommand) Run(args []string) int {
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
	secret, err := client.Logical().Read(path)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading %s: %s", path, err))
		return 2
	}
	if secret == nil {
		c.UI.Error(fmt.Sprintf("No value found at %s", path))
		return 2
	}

	if c.flagField != "" {
		return PrintRawField(c.UI, secret, c.flagField)
	}

	// If we have wrap info print the secret normally.
	if secret.WrapInfo != nil || c.flagFormat != "table" {
		return OutputSecret(c.UI, secret)
	}

	versionsRaw, ok := secret.Data["versions"]
	if !ok || versionsRaw == nil {
		c.UI.Error(fmt.Sprintf("No value found at %s", path))
		OutputSecret(c.UI, secret)
		return 2
	}
	versions := versionsRaw.(map[string]interface{})

	delete(secret.Data, "versions")

	c.UI.Info(getHeaderForMap("Metadata", secret.Data))
	OutputSecret(c.UI, secret)

	var versionKeys []int
	for k := range versions {
		i, err := strconv.Atoi(k)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error parsing version %s", k))
			return 2
		}

		versionKeys = append(versionKeys, i)
	}

	sort.Ints(versionKeys)

	for _, v := range versionKeys {
		c.UI.Info("\n" + getHeaderForMap(fmt.Sprintf("Version %d", v), versions[strconv.Itoa(v)].(map[string]interface{})))
		OutputData(c.UI, versions[strconv.Itoa(v)])
	}

	return 0
}
