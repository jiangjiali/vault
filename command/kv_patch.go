package command

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVPatchCommand)(nil)
var _ cli.CommandAutocomplete = (*KVPatchCommand)(nil)

type KVPatchCommand struct {
	*BaseCommand

	testStdin io.Reader // for tests
}

func (c *KVPatchCommand) Synopsis() string {
	return "在不覆盖的情况下设置或更新KV存储中的数据"
}

func (c *KVPatchCommand) Help() string {
	helpText := `
使用: vault kv patch [选项] KEY [DATA]

 *注意*：这只支持KV v2机密引擎。

  将数据写入键值存储区中的给定路径。数据可以是任何类型。

      $ vault kv patch secret/foo bar=baz

  数据也可以从磁盘上的文件中通过前缀“@”符号来使用。例如：

      $ vault kv patch secret/foo @data.json

  也可以使用“-”符号从stdin读取：

      $ echo "abcd1234" | vault kv patch secret/foo bar=-

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *KVPatchCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)

	return set
}

func (c *KVPatchCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *KVPatchCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVPatchCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	// Pull our fake stdin if needed
	stdin := (io.Reader)(os.Stdin)
	if c.testStdin != nil {
		stdin = c.testStdin
	}

	switch {
	case len(args) < 1:
		c.UI.Error(fmt.Sprintf("Not enough arguments (expected >1, got %d)", len(args)))
		return 1
	case len(args) == 1:
		c.UI.Error("Must supply data")
		return 1
	}

	var err error
	path := sanitizePath(args[0])

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	newData, err := parseArgsData(stdin, args[1:])
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to parse K=V data: %s", err))
		return 1
	}

	mountPath, v2, err := isKVv2(path, client)
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	if !v2 {
		c.UI.Error(fmt.Sprintf("K/V engine mount must be version 2 for patch support"))
		return 2
	}

	path = addPrefixToVKVPath(path, mountPath, "data")

	// First, do a read
	secret, err := kvReadRequest(client, path, nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error doing pre-read at %s: %s", path, err))
		return 2
	}

	// Make sure a value already exists
	if secret == nil || secret.Data == nil {
		c.UI.Error(fmt.Sprintf("No value found at %s", path))
		return 2
	}

	// Verify metadata found
	rawMeta, ok := secret.Data["metadata"]
	if !ok || rawMeta == nil {
		c.UI.Error(fmt.Sprintf("No metadata found at %s; patch only works on existing data", path))
		return 2
	}
	meta, ok := rawMeta.(map[string]interface{})
	if !ok {
		c.UI.Error(fmt.Sprintf("Metadata found at %s is not the expected type (JSON object)", path))
		return 2
	}
	if meta == nil {
		c.UI.Error(fmt.Sprintf("No metadata found at %s; patch only works on existing data", path))
		return 2
	}

	// Verify old data found
	rawData, ok := secret.Data["data"]
	if !ok || rawData == nil {
		c.UI.Error(fmt.Sprintf("No data found at %s; patch only works on existing data", path))
		return 2
	}
	data, ok := rawData.(map[string]interface{})
	if !ok {
		c.UI.Error(fmt.Sprintf("Data found at %s is not the expected type (JSON object)", path))
		return 2
	}
	if data == nil {
		c.UI.Error(fmt.Sprintf("No data found at %s; patch only works on existing data", path))
		return 2
	}

	// Copy new data over
	for k, v := range newData {
		data[k] = v
	}

	secret, err = client.Logical().Write(path, map[string]interface{}{
		"data": data,
		"options": map[string]interface{}{
			"cas": meta["version"],
		},
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error writing data to %s: %s", path, err))
		return 2
	}
	if secret == nil {
		// Don't output anything unless using the "table" format
		if Format(c.UI) == "table" {
			c.UI.Info(fmt.Sprintf("Success! Data written to: %s", path))
		}
		return 0
	}

	if c.flagField != "" {
		return PrintRawField(c.UI, secret, c.flagField)
	}

	return OutputSecret(c.UI, secret)
}
