package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/consts"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PluginDeregisterCommand)(nil)
var _ cli.CommandAutocomplete = (*PluginDeregisterCommand)(nil)

type PluginDeregisterCommand struct {
	*BaseCommand
}

func (c *PluginDeregisterCommand) Synopsis() string {
	return "注销目录中的现有插件"
}

func (c *PluginDeregisterCommand) Help() string {
	helpText := `
使用: vault plugin deregister [选项] TYPE NAME

  注销目录中的现有插件。如果插件不存在，则不执行任何操作（命令为幂等）。
  类型的参数采用“auth”、“database”或“secret”。

  注销名为“my-custom-plugin”的插件：

      $ vault plugin deregister auth my-custom-plugin

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PluginDeregisterCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP)
}

func (c *PluginDeregisterCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultPlugins(consts.PluginTypeUnknown)
}

func (c *PluginDeregisterCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PluginDeregisterCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	var pluginNameRaw, pluginTypeRaw string
	args = f.Args()
	switch {
	case len(args) < 1:
		c.UI.Error(fmt.Sprintf("Not enough arguments (expected 1 or 2, got %d)", len(args)))
		return 1
	case len(args) > 2:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1 or 2, got %d)", len(args)))
		return 1

	// These cases should come after invalid cases have been checked
	case len(args) == 1:
		pluginTypeRaw = "unknown"
		pluginNameRaw = args[0]
	case len(args) == 2:
		pluginTypeRaw = args[0]
		pluginNameRaw = args[1]
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	pluginType, err := consts.ParsePluginType(strings.TrimSpace(pluginTypeRaw))
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}
	pluginName := strings.TrimSpace(pluginNameRaw)

	if err := client.Sys().DeregisterPlugin(&api.DeregisterPluginInput{
		Name: pluginName,
		Type: pluginType,
	}); err != nil {
		c.UI.Error(fmt.Sprintf("Error deregistering plugin named %s: %s", pluginName, err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Deregistered plugin (if it was registered): %s", pluginName))
	return 0
}
