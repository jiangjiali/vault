package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/consts"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PluginInfoCommand)(nil)
var _ cli.CommandAutocomplete = (*PluginInfoCommand)(nil)

type PluginInfoCommand struct {
	*BaseCommand
}

func (c *PluginInfoCommand) Synopsis() string {
	return "阅读目录中有关插件的信息"
}

func (c *PluginInfoCommand) Help() string {
	helpText := `
使用: vault plugin info [选项] TYPE NAME

  显示目录中具有给定名称的插件的信息。如果插件不存在，则返回一个错误。
  类型的参数采用“auth”、“database”或“secret”。

  获取有关插件的信息：

      $ vault plugin info database mysql-database-plugin

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PluginInfoCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)
}

func (c *PluginInfoCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultPlugins(consts.PluginTypeUnknown)
}

func (c *PluginInfoCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PluginInfoCommand) Run(args []string) int {
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

	resp, err := client.Sys().GetPlugin(&api.GetPluginInput{
		Name: pluginName,
		Type: pluginType,
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading plugin named %s: %s", pluginName, err))
		return 2
	}

	if resp == nil {
		c.UI.Error(fmt.Sprintf("No value found for plugin %q", pluginName))
		return 2
	}

	data := map[string]interface{}{
		"args":    resp.Args,
		"builtin": resp.Builtin,
		"command": resp.Command,
		"name":    resp.Name,
		"sha256":  resp.SHA256,
	}

	if c.flagField != "" {
		return PrintRawField(c.UI, data, c.flagField)
	}
	return OutputData(c.UI, data)
}
