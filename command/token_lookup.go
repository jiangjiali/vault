package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*TokenLookupCommand)(nil)
var _ cli.CommandAutocomplete = (*TokenLookupCommand)(nil)

type TokenLookupCommand struct {
	*BaseCommand

	flagAccessor bool
}

func (c *TokenLookupCommand) Synopsis() string {
	return "显示有关令牌的信息"
}

func (c *TokenLookupCommand) Help() string {
	helpText := `
使用: vault token lookup [选项] [TOKEN | ACCESSOR]

  显示有关令牌或访问器的信息。如果没有提供令牌，则使用本地认证令牌。

  获取有关本地身份验证令牌的信息
  （这将使用“/auth/token/lookup-self”端点和权限）：

      $ vault token lookup

  获取有关特定令牌的信息
  （这将使用“/auth/token/lookup”端点和权限）：

      $ vault token lookup 96ddf4bc-d217-f3ba-f9bd-017055595017

  通过令牌的访问器获取有关令牌的信息：

      $ vault token lookup -accessor 9793c9b3-e04a-46f3-e7b8-748d7da248da

  有关示例的完整列表，请参阅文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *TokenLookupCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:       "accessor",
		Target:     &c.flagAccessor,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "将参数视为访问器而不是令牌。选择此选项时，输出将不包括令牌。",
	})

	return set
}

func (c *TokenLookupCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFiles()
}

func (c *TokenLookupCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *TokenLookupCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	token := ""

	args = f.Args()
	switch {
	case c.flagAccessor && len(args) < 1:
		c.UI.Error(fmt.Sprintf("Not enough arguments with -accessor (expected 1, got %d)", len(args)))
		return 1
	case c.flagAccessor && len(args) > 1:
		c.UI.Error(fmt.Sprintf("Too many arguments with -accessor (expected 1, got %d)", len(args)))
		return 1
	case len(args) == 0:
		// Use the local token
	case len(args) == 1:
		token = strings.TrimSpace(args[0])
	case len(args) > 1:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 0-1, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	var secret *api.Secret
	switch {
	case token == "":
		secret, err = client.Auth().Token().LookupSelf()
	case c.flagAccessor:
		secret, err = client.Auth().Token().LookupAccessor(token)
	default:
		secret, err = client.Auth().Token().Lookup(token)
	}

	if err != nil {
		c.UI.Error(fmt.Sprintf("Error looking up token: %s", err))
		return 2
	}

	return OutputSecret(c.UI, secret)
}
