package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*OperatorStepDownCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorStepDownCommand)(nil)

type OperatorStepDownCommand struct {
	*BaseCommand
}

func (c *OperatorStepDownCommand) Synopsis() string {
	return "迫使安全库辞去现役"
}

func (c *OperatorStepDownCommand) Help() string {
	helpText := `
使用: vault operator step-down [选项]

  强制位于给定地址的保管库服务器退出现役。虽然受影响的节点在再次尝试
  获取引线锁之前会有一段延迟，但是如果没有其他安全库节点预先获取锁，
  则同一节点可能会重新获取锁并再次激活。

  迫使安全库辞去领导职务：

      $ vault operator step-down

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorStepDownCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *OperatorStepDownCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *OperatorStepDownCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorStepDownCommand) Run(args []string) int {
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

	if err := client.Sys().StepDown(); err != nil {
		c.UI.Error(fmt.Sprintf("Error stepping down: %s", err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Stepped down: %s", client.Address()))
	return 0
}
