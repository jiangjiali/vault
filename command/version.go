package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/version"
)

var _ cli.Command = (*VersionCommand)(nil)
var _ cli.CommandAutocomplete = (*VersionCommand)(nil)

// VersionCommand is a Command implementation prints the version.
type VersionCommand struct {
	*BaseCommand

	VersionInfo *version.VInfo
}

func (c *VersionCommand) Synopsis() string {
	return "打印版本号"
}

func (c *VersionCommand) Help() string {
	helpText := `
使用: vault version

  打印此安全库CLI的版本。这不会打印目标安全库服务器版本。

  打印版本号：

      $ vault version

  此命令没有参数或标志。忽略任何其他参数或标志。
`
	return strings.TrimSpace(helpText)
}

func (c *VersionCommand) Flags() *FlagSets {
	return nil
}

func (c *VersionCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *VersionCommand) AutocompleteFlags() complete.Flags {
	return nil
}

func (c *VersionCommand) Run(_ []string) int {
	out := c.VersionInfo.FullVersionNumber(true)
	if version.CgoEnabled {
		out += " (cgo)"
	}
	c.UI.Output(out)
	return 0
}
