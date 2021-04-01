package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*TokenRevokeCommand)(nil)
var _ cli.CommandAutocomplete = (*TokenRevokeCommand)(nil)

type TokenRevokeCommand struct {
	*BaseCommand

	flagAccessor bool
	flagSelf     bool
	flagMode     string
}

func (c *TokenRevokeCommand) Synopsis() string {
	return "撤消令牌及其子代"
}

func (c *TokenRevokeCommand) Help() string {
	helpText := `
使用: vault token revoke [选项] [TOKEN | ACCESSOR]

  撤消身份验证令牌及其子级。如果没有提供令牌，则使用本地认证令牌。
  “-mode”标志可用于控制吊销的行为。有关详细信息，请参阅“-mode”标志文档。

  吊销令牌及其所有子代：

      $ vault token revoke 96ddf4bc-d217-f3ba-f9bd-017055595017

  撤消一个离开令牌子代的令牌：

      $ vault token revoke -mode=orphan 96ddf4bc-d217-f3ba-f9bd-017055595017

  由访问者吊销令牌：

      $ vault token revoke -accessor 9793c9b3-e04a-46f3-e7b8-748d7da248da

  有关示例的完整列表，请参阅文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *TokenRevokeCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP)

	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:       "accessor",
		Target:     &c.flagAccessor,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "将参数视为访问器而不是令牌。",
	})

	f.BoolVar(&BoolVar{
		Name:       "self",
		Target:     &c.flagSelf,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "对当前经过身份验证的令牌执行吊销。",
	})

	f.StringVar(&StringVar{
		Name:       "mode",
		Target:     &c.flagMode,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictSet("orphan", "path"),
		Usage:      "要执行的吊销类型。如果未指定，安全库将吊销令牌及其所有子代。如果是“orphan”，金库将只撤销令牌，将孩子们作为孤儿。如果为“path”，则从给定的身份验证路径前缀创建的令牌将与其子级一起删除。",
	})

	return set
}

func (c *TokenRevokeCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *TokenRevokeCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *TokenRevokeCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	token := ""
	if len(args) > 0 {
		token = strings.TrimSpace(args[0])
	}

	switch c.flagMode {
	case "", "orphan", "path":
	default:
		c.UI.Error(fmt.Sprintf("Invalid mode: %s", c.flagMode))
		return 1
	}

	switch {
	case c.flagSelf && len(args) > 0:
		c.UI.Error(fmt.Sprintf("Too many arguments with -self (expected 0, got %d)", len(args)))
		return 1
	case !c.flagSelf && len(args) > 1:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1 or -self, got %d)", len(args)))
		return 1
	case !c.flagSelf && len(args) < 1:
		c.UI.Error(fmt.Sprintf("Not enough arguments (expected 1 or -self, got %d)", len(args)))
		return 1
	case c.flagSelf && c.flagAccessor:
		c.UI.Error("Cannot use -self with -accessor!")
		return 1
	case c.flagSelf && c.flagMode != "":
		c.UI.Error("Cannot use -self with -mode!")
		return 1
	case c.flagAccessor && c.flagMode == "orphan":
		c.UI.Error("Cannot use -accessor with -mode=orphan!")
		return 1
	case c.flagAccessor && c.flagMode == "path":
		c.UI.Error("Cannot use -accessor with -mode=path!")
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	var revokeFn func(string) error
	// Handle all 6 possible combinations
	switch {
	case !c.flagAccessor && c.flagSelf && c.flagMode == "":
		revokeFn = client.Auth().Token().RevokeSelf
	case !c.flagAccessor && !c.flagSelf && c.flagMode == "":
		revokeFn = client.Auth().Token().RevokeTree
	case !c.flagAccessor && !c.flagSelf && c.flagMode == "orphan":
		revokeFn = client.Auth().Token().RevokeOrphan
	case !c.flagAccessor && !c.flagSelf && c.flagMode == "path":
		revokeFn = client.Sys().RevokePrefix
	case c.flagAccessor && !c.flagSelf && c.flagMode == "":
		revokeFn = client.Auth().Token().RevokeAccessor
	}

	if err := revokeFn(token); err != nil {
		c.UI.Error(fmt.Sprintf("Error revoking token: %s", err))
		return 2
	}

	c.UI.Output("Success! Revoked token (if it existed)")
	return 0
}
