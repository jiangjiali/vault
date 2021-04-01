package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*OperatorSealCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorSealCommand)(nil)

type OperatorSealCommand struct {
	*BaseCommand
}

func (c *OperatorSealCommand) Synopsis() string {
	return "密封服务器"
}

func (c *OperatorSealCommand) Help() string {
	helpText := `
使用: vault operator seal [选项]

  密封安全库服务器。密封告诉安全库服务器在解封之前停止响应任何操作。
  密封后，安全库服务器将丢弃其内存中的主密钥以解锁数据，
  因此它将被物理阻止对未密封操作的响应。

  如果正在进行解封，则密封安全库将重置解封过程。
  用户必须重新输入他们的主密钥部分。

  如果安全库服务器已密封，则此命令不执行任何操作。

  密封安全库服务器：

      $ vault operator seal

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorSealCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *OperatorSealCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *OperatorSealCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorSealCommand) Run(args []string) int {
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

	if err := client.Sys().Seal(); err != nil {
		c.UI.Error(fmt.Sprintf("Error sealing: %s", err))
		return 2
	}

	c.UI.Output("Success! Vault is sealed.")
	return 0
}
