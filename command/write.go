package command

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*WriteCommand)(nil)
var _ cli.CommandAutocomplete = (*WriteCommand)(nil)

// WriteCommand is a Command that puts data into the Vault.
type WriteCommand struct {
	*BaseCommand

	flagForce bool

	testStdin io.Reader // for tests
}

func (c *WriteCommand) Synopsis() string {
	return "写入数据、配置和机密"
}

func (c *WriteCommand) Help() string {
	helpText := `
使用: vault write [选项] PATH [DATA K=V...]

  将数据写入给定路径的安全库。数据可以是凭证、机密、配置或任意数据。
  这个命令的具体行为是由路径上安装的东西决定的。

  数据被指定为“key=value”对。如果值以“@”开头，则从文件加载。
  如果值为“-”，则安全库将从stdin读取该值。

  在通用机密引擎中保存数据：

      $ vault write secret/my-secret foo=bar

  在传输机密引擎中创建新的加密密钥：

      $ vault write -f transit/keys/my-key

  有关示例和路径的完整列表，请参阅与所使用的秘密引擎相对应的文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *WriteCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)
	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:       "force",
		Aliases:    []string{"f"},
		Target:     &c.flagForce,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "允许操作在没有key=value对的情况下继续。这允许写入不需要或不需要数据的键。",
	})

	return set
}

func (c *WriteCommand) AutocompleteArgs() complete.Predictor {
	// Return an anything predictor here. Without a way to access help
	// information, we don't know what paths we could write to.
	return complete.PredictAnything
}

func (c *WriteCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *WriteCommand) Run(args []string) int {
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
	case len(args) == 1 && !c.flagForce:
		c.UI.Error("Must supply data or use -force")
		return 1
	}

	// Pull our fake stdin if needed
	stdin := (io.Reader)(os.Stdin)
	if c.testStdin != nil {
		stdin = c.testStdin
	}

	path := sanitizePath(args[0])

	data, err := parseArgsData(stdin, args[1:])
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to parse K=V data: %s", err))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
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

	// Handle single field output
	if c.flagField != "" {
		return PrintRawField(c.UI, secret, c.flagField)
	}

	return OutputSecret(c.UI, secret)
}
