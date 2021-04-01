package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*StatusCommand)(nil)
var _ cli.CommandAutocomplete = (*StatusCommand)(nil)

type StatusCommand struct {
	*BaseCommand
}

func (c *StatusCommand) Synopsis() string {
	return "打印密封和HA状态"
}

func (c *StatusCommand) Help() string {
	helpText := `
使用: vault status [选项]

  打印安全库的当前状态，包括是否已密封以及是否启用HA模式。
  无论安全库是否密封，此命令都会打印。

  出口代码反映封条状态：

      - 0 - 未密封
      - 1 - 错误
      - 2 - 密封

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *StatusCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *StatusCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *StatusCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *StatusCommand) Run(args []string) int {
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
		// We return 2 everywhere else, but 2 is reserved for "sealed" here
		return 1
	}

	status, err := client.Sys().SealStatus()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error checking seal status: %s", err))
		return 1
	}

	// Do not return the int here yet, since we may want to return a custom error
	// code depending on the seal status.
	code := OutputSealStatus(c.UI, client, status)

	if status.Sealed {
		return 2
	}

	return code
}
