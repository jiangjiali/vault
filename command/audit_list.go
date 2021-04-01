package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"sort"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*AuditListCommand)(nil)
var _ cli.CommandAutocomplete = (*AuditListCommand)(nil)

type AuditListCommand struct {
	*BaseCommand

	flagDetailed bool
}

func (c *AuditListCommand) Synopsis() string {
	return "列出已启用的审核设备"
}

func (c *AuditListCommand) Help() string {
	helpText := `
使用: vault audit list [选项]

  列出Vault服务器中已启用的审核设备。
  输出将列出已启用的审核设备以及这些设备的选项。

  列出所有审核设备：

      $ vault audit list

  列出有关审核设备的详细输出：

      $ vault audit list -detailed

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *AuditListCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:    "detailed",
		Target:  &c.flagDetailed,
		Default: false,
		EnvVar:  "",
		Usage:   "打印有关每个身份验证设备的选项和复制状态等详细信息。",
	})

	return set
}

func (c *AuditListCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *AuditListCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *AuditListCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	if len(args) > 0 {
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 0, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	audits, err := client.Sys().ListAudit()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error listing audits: %s", err))
		return 2
	}

	switch Format(c.UI) {
	case "table":
		if len(audits) == 0 {
			c.UI.Output(fmt.Sprintf("No audit devices are enabled."))
			return 2
		}

		if c.flagDetailed {
			c.UI.Output(tableOutput(c.detailedAudits(audits), nil))
			return 0
		}
		c.UI.Output(tableOutput(c.simpleAudits(audits), nil))
		return 0
	default:
		return OutputData(c.UI, audits)
	}
}

func (c *AuditListCommand) simpleAudits(audits map[string]*api.Audit) []string {
	paths := make([]string, 0, len(audits))
	for path := range audits {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	columns := []string{"Path | Type | Description"}
	for _, path := range paths {
		audit := audits[path]
		columns = append(columns, fmt.Sprintf("%s | %s | %s",
			audit.Path,
			audit.Type,
			audit.Description,
		))
	}

	return columns
}

func (c *AuditListCommand) detailedAudits(audits map[string]*api.Audit) []string {
	paths := make([]string, 0, len(audits))
	for path := range audits {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	columns := []string{"Path | Type | Description | Replication | Options"}
	for _, path := range paths {
		audit := audits[path]

		opts := make([]string, 0, len(audit.Options))
		for k, v := range audit.Options {
			opts = append(opts, k+"="+v)
		}

		replication := "replicated"
		if audit.Local {
			replication = "local"
		}

		columns = append(columns, fmt.Sprintf("%s | %s | %s | %s | %s",
			path,
			audit.Type,
			audit.Description,
			replication,
			strings.Join(opts, " "),
		))
	}

	return columns
}
