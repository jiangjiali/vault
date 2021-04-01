package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*PluginCommand)(nil)

type PluginCommand struct {
	*BaseCommand
}

func (c *PluginCommand) Synopsis() string {
	return "与插件和目录交互"
}

func (c *PluginCommand) Help() string {
	helpText := `
使用: vault plugin <子命令> [选项] [参数]

  此命令将用于与安全库插件和插件目录交互的子命令分组。
  插件目录分为三种类型：“auth”、“database”和“secret”插件。
  每次调用都必须指定类型。下面是一些插件命令的示例。

  列出特定类型目录中的所有可用插件：

      $ vault plugin list database

  将新插件注册为特定类型：

      $ vault plugin register -sha256=d3f0a8b... auth my-custom-plugin

  获取目录中特定类型下列出的插件的信息：

      $ vault plugin info auth my-custom-plugin

  重新加载插件后端：

      $ vault plugin reload -plugin=my-custom-plugin

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *PluginCommand) Run(args []string) int {
	return cli.RunResultHelp
}
