package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/consts"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PluginRegisterCommand)(nil)
var _ cli.CommandAutocomplete = (*PluginRegisterCommand)(nil)

type PluginRegisterCommand struct {
	*BaseCommand

	flagArgs    []string
	flagCommand string
	flagSHA256  string
}

func (c *PluginRegisterCommand) Synopsis() string {
	return "在目录中注册新插件"
}

func (c *PluginRegisterCommand) Help() string {
	helpText := `
使用: vault plugin register [选项] TYPE NAME

  在目录中注册一个新插件。插件二进制文件必须存在于安全库配置的插件目录中。
  类型的参数采用“auth”或“secret”。

  注册名为my-custom-plugin的插件：

      $ vault plugin register -sha256=d3f0a8b... auth my-custom-plugin

  使用自定义参数注册插件：

      $ vault plugin register \
          -sha256=d3f0a8b... \
          -args=--with-glibc,--with-cgo \
          auth my-custom-plugin

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PluginRegisterCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP)

	f := set.NewFlagSet("命令选项")

	f.StringSliceVar(&StringSliceVar{
		Name:       "args",
		Target:     &c.flagArgs,
		Completion: complete.PredictAnything,
		Usage:      "启动时传递给插件的参数。用逗号分隔多个参数。",
	})

	f.StringVar(&StringVar{
		Name:       "command",
		Target:     &c.flagCommand,
		Completion: complete.PredictAnything,
		Usage:      "生成插件的命令。如果未指定，则默认为插件的名称。",
	})

	f.StringVar(&StringVar{
		Name:       "sha256",
		Target:     &c.flagSHA256,
		Completion: complete.PredictAnything,
		Usage:      "插件二进制文件的SHA256。这是所有插件所必需的。",
	})

	return set
}

func (c *PluginRegisterCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultPlugins(consts.PluginTypeUnknown)
}

func (c *PluginRegisterCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PluginRegisterCommand) Run(args []string) int {
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
	case c.flagSHA256 == "":
		c.UI.Error("SHA256 is required for all plugins, please provide -sha256")
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

	command := c.flagCommand
	if command == "" {
		command = pluginName
	}

	if err := client.Sys().RegisterPlugin(&api.RegisterPluginInput{
		Name:    pluginName,
		Type:    pluginType,
		Args:    c.flagArgs,
		Command: command,
		SHA256:  c.flagSHA256,
	}); err != nil {
		c.UI.Error(fmt.Sprintf("Error registering plugin %s: %s", pluginName, err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Registered plugin: %s", pluginName))
	return 0
}
