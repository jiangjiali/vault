package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/api"
	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PluginReloadCommand)(nil)
var _ cli.CommandAutocomplete = (*PluginReloadCommand)(nil)

type PluginReloadCommand struct {
	*BaseCommand
	plugin string
	mounts []string
	scope  string
}

func (c *PluginReloadCommand) Synopsis() string {
	return "重新加载插件后端"
}

func (c *PluginReloadCommand) Help() string {
	helpText := `
使用: vault plugin reload [选项]

  重新加载挂载的插件。必须提供插件名称或所需的插件装载，
  但不能同时提供两者。如果提供了插件名称，
  则将重新加载使用插件后端的所有相应挂载路径。

  重新加载名为 "my-custom-plugin" 的插件:

	  $ vault plugin reload -plugin=my-custom-plugin

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PluginReloadCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP)

	f := set.NewFlagSet("选项")

	f.StringVar(&StringVar{
		Name:       "plugin",
		Target:     &c.plugin,
		Completion: complete.PredictAnything,
		Usage:      "要重新加载的插件的名称，在插件目录中注册。",
	})

	f.StringSliceVar(&StringSliceVar{
		Name:       "mounts",
		Target:     &c.mounts,
		Completion: complete.PredictAnything,
		Usage:      "要重新加载的插件后端的数组或逗号分隔的字符串安装路径。",
	})

	f.StringVar(&StringVar{
		Name:       "scope",
		Target:     &c.scope,
		Completion: complete.PredictAnything,
		Usage:      "重新加载的范围，对于“local”，“global”，对于复制的重新加载，省略了。",
	})

	return set
}

func (c *PluginReloadCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *PluginReloadCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PluginReloadCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	switch {
	case c.plugin == "" && len(c.mounts) == 0:
		c.UI.Error(fmt.Sprintf("Not enough arguments (expected 1, got %d)", len(args)))
		return 1
	case c.plugin != "" && len(c.mounts) > 0:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	case c.scope != "" && c.scope != "global":
		c.UI.Error(fmt.Sprintf("无效的重新加载的范围: %s", c.scope))
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	rid, err := client.Sys().ReloadPlugin(&api.ReloadPluginInput{
		Plugin: c.plugin,
		Mounts: c.mounts,
		Scope:  c.scope,
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("重新加载插件/挂载时出错: %s", err))
		return 2
	}

	if len(c.mounts) > 0 {
		if rid != "" {
			c.UI.Output(fmt.Sprintf("Success! Reloading mounts: %s, reload_id: %s", c.mounts, rid))
		} else {
			c.UI.Output(fmt.Sprintf("Success! Reloaded mounts: %s", c.mounts))
		}
	} else {
		if rid != "" {
			c.UI.Output(fmt.Sprintf("Success! Reloading plugin: %s, reload_id: %s", c.plugin, rid))
		} else {
			c.UI.Output(fmt.Sprintf("Success! Reloaded plugin: %s", c.plugin))
		}
	}

	return 0
}
