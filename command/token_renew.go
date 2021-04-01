package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"
	"time"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*TokenRenewCommand)(nil)
var _ cli.CommandAutocomplete = (*TokenRenewCommand)(nil)

type TokenRenewCommand struct {
	*BaseCommand

	flagIncrement time.Duration
}

func (c *TokenRenewCommand) Synopsis() string {
	return "续订令牌租约"
}

func (c *TokenRenewCommand) Help() string {
	helpText := `
使用: vault token renew [选项] [TOKEN]

  续订令牌的租约，延长其可使用的时间。如果没有提供令牌，
  则使用本地认证令牌。如果令牌不可续订、令牌已被吊销
  或令牌已达到其最大TTL，则租约续订将失败。

  续订令牌（这将使用“/auth/token/renew”端点和权限）：

      $ vault token renew 96ddf4bc-d217-f3ba-f9bd-017055595017

  续订当前已验证的令牌（这将使用“/auth/token/renew-self”端点和权限）：

      $ vault token renew

  续订请求特定增量值的令牌：

      $ vault token renew -increment=30m 96ddf4bc-d217-f3ba-f9bd-017055595017

  有关示例的完整列表，请参阅文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *TokenRenewCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
	f := set.NewFlagSet("命令选项")

	f.DurationVar(&DurationVar{
		Name:       "increment",
		Aliases:    []string{"i"},
		Target:     &c.flagIncrement,
		Default:    0,
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "申请续约的特定增量。安全库不需要满足此请求。如果未提供，则安全库将使用默认TTL。它被指定为带有后缀“30s”或“5m”的数字字符串。",
	})

	return set
}

func (c *TokenRenewCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFiles()
}

func (c *TokenRenewCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *TokenRenewCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	token := ""
	increment := c.flagIncrement

	args = f.Args()
	switch len(args) {
	case 0:
		// Use the local token
	case 1:
		token = strings.TrimSpace(args[0])
	default:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	var secret *api.Secret
	inc := truncateToSeconds(increment)
	if token == "" {
		secret, err = client.Auth().Token().RenewSelf(inc)
	} else {
		secret, err = client.Auth().Token().Renew(token, inc)
	}
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error renewing token: %s", err))
		return 2
	}

	return OutputSecret(c.UI, secret)
}
