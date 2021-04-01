package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*ListCommand)(nil)
var _ cli.CommandAutocomplete = (*ListCommand)(nil)

type ListCommand struct {
	*BaseCommand
}

func (c *ListCommand) Synopsis() string {
	return "列出数据或机密"
}

func (c *ListCommand) Help() string {
	helpText := `

使用: vault list [选项] PATH

  列出位于给定路径的服务器中的数据。
  这可以用来列出给定机密引擎中的密钥。

  在通用机密引擎的“my app”文件夹下列出值：

      $ vault list secret/my-app/

  有关示例和路径的完整列表，请参阅与所使用的秘密引擎对应的文档。
  并非所有引擎都支持列表。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *ListCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *ListCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFolders()
}

func (c *ListCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *ListCommand) Run(args []string) int {
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

	if secret == nil {
		c.UI.Error(fmt.Sprintf("No value found at %s", path))
		return 2
	}
	if secret.Data == nil {
		// If secret wasn't nil, we have warnings, so output them anyways. We
		// may also have non-keys info.
		return OutputSecret(c.UI, secret)
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
