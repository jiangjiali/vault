package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*OperatorRotateCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorRotateCommand)(nil)

type OperatorRotateCommand struct {
	*BaseCommand
}

func (c *OperatorRotateCommand) Synopsis() string {
	return "旋转基础加密密钥"
}

func (c *OperatorRotateCommand) Help() string {
	helpText := `
使用: vault operator rotate [选项]

  旋转用于保护写入存储后端的数据的基础加密密钥。
  这将在密钥环中安装新密钥。这个新密钥用于加密新数据，
  而环中的旧密钥用于解密旧数据。

  这是一个在线操作，不会导致停机。
  此命令针对每个群集（而不是每台服务器）运行，
  因为处于HA模式的安全库服务器共享同一个存储后端。

  旋转安全库的加密密钥：

      $ vault operator rotate

  有关示例的完整列表，请参阅文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorRotateCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *OperatorRotateCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *OperatorRotateCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorRotateCommand) Run(args []string) int {
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

	// Rotate the key
	err = client.Sys().Rotate()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error rotating key: %s", err))
		return 2
	}

	// Print the key status
	status, err := client.Sys().KeyStatus()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading key status: %s", err))
		return 2
	}

	switch Format(c.UI) {
	case "table":
		c.UI.Output("Success! Rotated key")
		c.UI.Output("")
		c.UI.Output(printKeyStatus(status))
		return 0
	default:
		return OutputData(c.UI, status)
	}
}
