package command

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVPutCommand)(nil)
var _ cli.CommandAutocomplete = (*KVPutCommand)(nil)

type KVPutCommand struct {
	*BaseCommand

	flagCAS   int
	testStdin io.Reader // for tests
}

func (c *KVPutCommand) Synopsis() string {
	return "设置或更新KV存储中的数据"
}

func (c *KVPutCommand) Help() string {
	helpText := `
使用: vault kv put [选项] KEY [DATA]

  将数据写入键值存储区中的给定路径。数据可以是任何类型。

      $ vault kv put secret/foo bar=baz

  数据也可以从磁盘上的文件中通过前缀“@”符号来使用。例如：

      $ vault kv put secret/foo @data.json

  也可以使用“-”符号从stdin读取：

      $ echo "abcd1234" | vault kv put secret/foo bar=-

  要执行检查和设置操作，请使用与要对其执行CAS操作的密钥
  对应的相应版本号指定-cas标志：

      $ vault kv put -cas=1 secret/foo bar=baz

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *KVPutCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)

	// Common Options
	f := set.NewFlagSet("命令选项")

	f.IntVar(&IntVar{
		Name:    "cas",
		Target:  &c.flagCAS,
		Default: -1,
		Usage:   `指定使用检查和设置操作。如果未设置，则允许写入。如果设置为0，则仅当密钥不存在时才允许写入。如果索引为非零，则只有当密钥的当前版本与cas参数中指定的版本匹配时，才允许写入。`,
	})

	return set
}

func (c *KVPutCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *KVPutCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVPutCommand) Run(args []string) int {
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

	data, err := parseArgsData(stdin, args[1:])
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to parse K=V data: %s", err))
		return 1
	}

	mountPath, v2, err := isKVv2(path, client)
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	if v2 {
		path = addPrefixToVKVPath(path, mountPath, "data")
		data = map[string]interface{}{
			"data":    data,
			"options": map[string]interface{}{},
		}

		if c.flagCAS > -1 {
			data["options"].(map[string]interface{})["cas"] = c.flagCAS
		}
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

	if c.flagField != "" {
		return PrintRawField(c.UI, secret, c.flagField)
	}

	return OutputSecret(c.UI, secret)
}
