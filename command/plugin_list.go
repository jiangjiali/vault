package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"sort"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/consts"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PluginListCommand)(nil)
var _ cli.CommandAutocomplete = (*PluginListCommand)(nil)

type PluginListCommand struct {
	*BaseCommand
}

func (c *PluginListCommand) Synopsis() string {
	return "列出可用的插件"
}

func (c *PluginListCommand) Help() string {
	helpText := `
使用: vault plugin list [选项] [TYPE]

  列出目录中注册的可用插件。这并没有列出插件是否在使用中，
  而是只列出它们的可用性。
  类型的最后一个参数采用“auth”、“database”或“secret”。

  列出目录中所有可用的插件：

      $ vault plugin list

  列出目录中所有可用的数据库插件：

      $ vault plugin list database

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PluginListCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetHTTP | FlagSetOutputFormat)
}

func (c *PluginListCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PluginListCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PluginListCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	switch {
	case len(args) > 1:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 0 or 1, got %d)", len(args)))
		return 1
	}

	pluginType := consts.PluginTypeUnknown
	if len(args) > 0 {
		pluginTypeStr := strings.TrimSpace(args[0])
		if pluginTypeStr != "" {
			var err error
			pluginType, err = consts.ParsePluginType(pluginTypeStr)
			if err != nil {
				c.UI.Error(fmt.Sprintf("Error parsing type: %s", err))
				return 2
			}
		}
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	resp, err := client.Sys().ListPlugins(&api.ListPluginsInput{
		Type: pluginType,
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error listing available plugins: %s", err))
		return 2
	}
	if resp == nil {
		c.UI.Error("No response from server when listing plugins")
		return 2
	}

	switch Format(c.UI) {
	case "table":
		var flattenedNames []string
		namesAdded := make(map[string]bool)
		for _, names := range resp.PluginsByType {
			for _, name := range names {
				if ok := namesAdded[name]; !ok {
					flattenedNames = append(flattenedNames, name)
					namesAdded[name] = true
				}
			}
			sort.Strings(flattenedNames)
		}
		list := append([]string{"Plugins"}, flattenedNames...)
		c.UI.Output(tableOutput(list, nil))
		return 0
	default:
		res := make(map[string]interface{})
		for k, v := range resp.PluginsByType {
			res[k.String()] = v
		}
		return OutputData(c.UI, res)
	}
}
