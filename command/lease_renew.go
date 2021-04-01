package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*LeaseRenewCommand)(nil)
var _ cli.CommandAutocomplete = (*LeaseRenewCommand)(nil)

type LeaseRenewCommand struct {
	*BaseCommand

	flagIncrement time.Duration
}

func (c *LeaseRenewCommand) Synopsis() string {
	return "续约机密"
}

func (c *LeaseRenewCommand) Help() string {
	helpText := `
使用: vault lease renew [选项] ID

  根据机密续订租约，延长在服务器吊销租约之前可以使用的时间。

  安全库里的每一个机密都有一份租约。
  如果这个机密的所有者想要使用它比租约更长的时间，那么它必须被更新。
  续约不会改变秘密的内容。ID是完整路径租约ID。

  更新机密：

      $ vault lease renew database/creds/readonly/2f6a614c...

  如果机密不可更新、机密已被吊销或机密已达到其最大TTL，则租约续订将失败。
  有关示例的完整列表，请参阅文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *LeaseRenewCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
	f := set.NewFlagSet("命令选项")

	f.DurationVar(&DurationVar{
		Name:       "increment",
		Target:     &c.flagIncrement,
		Default:    0,
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "请求以秒为单位的特定增量。安全库不需要满足此请求。",
	})

	return set
}

func (c *LeaseRenewCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictAnything
}

func (c *LeaseRenewCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *LeaseRenewCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	leaseID := ""
	increment := c.flagIncrement

	args = f.Args()
	switch len(args) {
	case 0:
		c.UI.Error("Missing ID!")
		return 1
	case 1:
		leaseID = strings.TrimSpace(args[0])
	default:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1-2, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	secret, err := client.Sys().Renew(leaseID, truncateToSeconds(increment))
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error renewing %s: %s", leaseID, err))
		return 2
	}

	return OutputSecret(c.UI, secret)
}
