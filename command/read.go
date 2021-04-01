package command

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*ReadCommand)(nil)
var _ cli.CommandAutocomplete = (*ReadCommand)(nil)

type ReadCommand struct {
	*BaseCommand

	testStdin io.Reader // for tests
}

func (c *ReadCommand) Synopsis() string {
	return "读取数据并检索机密"
}

func (c *ReadCommand) Help() string {
	helpText := `
使用: vault read [选项] PATH

  从给定路径的安全库中读取数据。
  它可用于读取机密、生成动态凭据、获取配置详细信息等。

  从静态机密引擎读取机密：

      $ vault read secret/my-secret

  有关示例和路径的完整列表，请参阅与所使用的机密引擎对应的文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *ReadCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)
}

func (c *ReadCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFiles()
}

func (c *ReadCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *ReadCommand) Run(args []string) int {
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
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	// Pull our fake stdin if needed
	stdin := (io.Reader)(os.Stdin)
	if c.testStdin != nil {
		stdin = c.testStdin
	}

	path := sanitizePath(args[0])

	data, err := parseArgsDataStringLists(stdin, args[1:])
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to parse K=V data: %s", err))
		return 1
	}

	secret, err := client.Logical().ReadWithData(path, data)
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

	return OutputSecret(c.UI, secret)
}
