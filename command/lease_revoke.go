package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"strings"
)

var _ cli.Command = (*LeaseRevokeCommand)(nil)
var _ cli.CommandAutocomplete = (*LeaseRevokeCommand)(nil)

type LeaseRevokeCommand struct {
	*BaseCommand

	flagForce  bool
	flagPrefix bool
	flagSync   bool
}

func (c *LeaseRevokeCommand) Synopsis() string {
	return "撤销租约和机密"
}

func (c *LeaseRevokeCommand) Help() string {
	helpText := `
使用: vault lease revoke [选项] ID

  按租约ID吊销机密。此命令可以基于路径匹配的前缀吊销单个机密或多个机密。

  不使用 -force 时的默认行为是异步撤消；
  服务器将对吊销进行排队，并在失败时继续尝试（包括跨重新启动）。
  -sync 标志可用于强制执行同步操作，但随后由调用方在失败时重试。
  强制模式始终同步运行。

  撤销单一租约：

      $ vault lease revoke database/creds/readonly/2f6a614c...

  吊销角色的所有租约：

      $ vault lease revoke -prefix userpass/creds/deploy

  强制从保管库删除租约，即使秘密引擎吊销失败：

      $ vault lease revoke -force -prefix creds

  有关示例和路径的完整列表，请参阅与所使用的秘密引擎对应的文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *LeaseRevokeCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP)
	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:    "force",
		Aliases: []string{"f"},
		Target:  &c.flagForce,
		Default: false,
		Usage:   "即使秘密引擎吊销失败，也要从保险库中删除租约。这适用于目标机密引擎中的机密被手动删除的恢复情况。如果指定了此标志，则还需要 -prefix",
	})

	f.BoolVar(&BoolVar{
		Name:    "prefix",
		Target:  &c.flagPrefix,
		Default: false,
		Usage:   "将ID视为前缀而不是确切的租约ID。这样可以同时吊销多个租约。",
	})

	f.BoolVar(&BoolVar{
		Name:    "sync",
		Target:  &c.flagSync,
		Default: false,
		Usage:   "强制执行同步操作；失败时由客户端重试。",
	})

	return set
}

func (c *LeaseRevokeCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFiles()
}

func (c *LeaseRevokeCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *LeaseRevokeCommand) Run(args []string) int {
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

	if c.flagForce && !c.flagPrefix {
		c.UI.Error("Specifying -force requires also specifying -prefix")
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	leaseID := strings.TrimSpace(args[0])

	revokeOpts := &api.RevokeOptions{
		LeaseID: leaseID,
		Force:   c.flagForce,
		Prefix:  c.flagPrefix,
		Sync:    c.flagSync,
	}

	if c.flagForce {
		c.UI.Warn(wrapAtLength("Warning! Force-removing leases can cause Vault " +
			"to become out of sync with secret engines!"))
	}

	err = client.Sys().RevokeWithOptions(revokeOpts)
	if err != nil {
		switch {
		case c.flagForce:
			c.UI.Error(fmt.Sprintf("Error force revoking leases with prefix %s: %s", leaseID, err))
			return 2
		case c.flagPrefix:
			c.UI.Error(fmt.Sprintf("Error revoking leases with prefix %s: %s", leaseID, err))
			return 2
		default:
			c.UI.Error(fmt.Sprintf("Error revoking lease %s: %s", leaseID, err))
			return 2
		}
	}

	if c.flagForce {
		c.UI.Output(fmt.Sprintf("Success! Force revoked any leases with prefix: %s", leaseID))
		return 0
	}

	if c.flagSync {
		if c.flagPrefix {
			c.UI.Output(fmt.Sprintf("Success! Revoked any leases with prefix: %s", leaseID))
			return 0
		}
		c.UI.Output(fmt.Sprintf("Success! Revoked lease: %s", leaseID))
		return 0
	}

	c.UI.Output("All revocation operations queued successfully!")
	return 0
}
