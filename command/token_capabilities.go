package command

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*TokenCapabilitiesCommand)(nil)
var _ cli.CommandAutocomplete = (*TokenCapabilitiesCommand)(nil)

type TokenCapabilitiesCommand struct {
	*BaseCommand
}

func (c *TokenCapabilitiesCommand) Synopsis() string {
	return "在PATH上打印令牌的功能"
}

func (c *TokenCapabilitiesCommand) Help() string {
	helpText := `
使用: vault token capabilities [选项] [TOKEN] PATH

  获取给定路径的令牌功能。如果将令牌作为参数提供，
  则使用“/sys/capabilities”端点和权限。如果没有提供令牌，
  “/sys/capabilities self”端点和权限将与本地认证令牌一起使用。

  列出“secret/foo”路径上本地令牌的功能：

      $ vault token capabilities secret/foo

  列出“public/foo”路径上令牌的功能：

      $ vault token capabilities 96ddf4bc-d217-f3ba-f9bd-017055595017 public/foo

  有关示例的完整列表，请参阅文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *TokenCapabilitiesCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *TokenCapabilitiesCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *TokenCapabilitiesCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *TokenCapabilitiesCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	token := ""
	path := ""
	args = f.Args()
	switch len(args) {
	case 0:
		c.UI.Error(fmt.Sprintf("Not enough arguments (expected 1-2, got 0)"))
		return 1
	case 1:
		path = args[0]
	case 2:
		token, path = args[0], args[1]
	default:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1-2, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	var capabilities []string
	if token == "" {
		capabilities, err = client.Sys().CapabilitiesSelf(path)
	} else {
		capabilities, err = client.Sys().Capabilities(token, path)
	}
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error listing capabilities: %s", err))
		return 2
	}
	if capabilities == nil {
		c.UI.Error(fmt.Sprintf("No capabilities found"))
		return 1
	}

	switch Format(c.UI) {
	case "table":
		sort.Strings(capabilities)
		c.UI.Output(strings.Join(capabilities, ", "))
		return 0
	default:
		return OutputData(c.UI, capabilities)
	}
}
