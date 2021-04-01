package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*SecretsMoveCommand)(nil)
var _ cli.CommandAutocomplete = (*SecretsMoveCommand)(nil)

type SecretsMoveCommand struct {
	*BaseCommand
}

func (c *SecretsMoveCommand) Synopsis() string {
	return "把机密引擎移到新的路径上"
}

func (c *SecretsMoveCommand) Help() string {
	helpText := `
使用: vault secrets move [选项] SOURCE DESTINATION

  将现有机密引擎移动到新路径。旧机密引擎的任何租约都将被吊销，
  但与引擎关联的所有配置都将保留。

  此命令仅在命名空间中工作；不能用于将引擎移动到不同的命名空间。

  警告！移动一个现有的秘密引擎将撤销旧引擎的所有租约。

  将位于 secret/ 的现有机密引擎移动到 generic/：

      $ vault secrets move secret/ generic/

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *SecretsMoveCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *SecretsMoveCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultMounts()
}

func (c *SecretsMoveCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *SecretsMoveCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	switch {
	case len(args) < 2:
		c.UI.Error(fmt.Sprintf("Not enough arguments (expected 2, got %d)", len(args)))
		return 1
	case len(args) > 2:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 2, got %d)", len(args)))
		return 1
	}

	// Grab the source and destination
	source := ensureTrailingSlash(args[0])
	destination := ensureTrailingSlash(args[1])

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	if err := client.Sys().Remount(source, destination); err != nil {
		c.UI.Error(fmt.Sprintf("Error moving secrets engine %s to %s: %s", source, destination, err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Moved secrets engine %s to: %s", source, destination))
	return 0
}
