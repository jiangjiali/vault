package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*AuditCommand)(nil)

type AuditCommand struct {
	*BaseCommand
}

func (c *AuditCommand) Synopsis() string {
	return "审核设备交互"
}

func (c *AuditCommand) Help() string {
	helpText := `
使用: vault audit <子命令> [选项] [参数]

  此命令将子命令分组以与Vault的审核设备交互。
  用户可以列出、启用和禁用审核设备。

  列出所有启用的审核设备：

      $ vault audit list

  启用新的审计设备"file";

       $ vault audit enable file file_path=/var/log/audit.log

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *AuditCommand) Run(args []string) int {
	return cli.RunResultHelp
}
