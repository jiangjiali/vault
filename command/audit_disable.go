package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*AuditDisableCommand)(nil)
var _ cli.CommandAutocomplete = (*AuditDisableCommand)(nil)

type AuditDisableCommand struct {
	*BaseCommand
}

func (c *AuditDisableCommand) Synopsis() string {
	return "禁用审核设备"
}

func (c *AuditDisableCommand) Help() string {
	helpText := `
使用: vault audit disable [选项] PATH

  禁用审核设备。一旦禁用了某个审核设备，则不会向该设备发送将来的审核日志。
  与审核设备关联的数据不受影响。

  参数对应于审核设备的 PATH，而不是 TYPE

  禁用在启用的审核设备"file/":

      $ vault audit disable file/

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *AuditDisableCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *AuditDisableCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultAudits()
}

func (c *AuditDisableCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *AuditDisableCommand) Run(args []string) int {
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

	path := ensureTrailingSlash(sanitizePath(args[0]))

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	if err := client.Sys().DisableAudit(path); err != nil {
		c.UI.Error(fmt.Sprintf("Error disabling audit device: %s", err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Disabled audit device (if it was enabled) at: %s", path))

	return 0
}
