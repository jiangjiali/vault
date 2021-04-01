package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*FileCommand)(nil)

type FileCommand struct {
	*BaseCommand
}

func (c *FileCommand) Synopsis() string {
	return "本地文件加密和解密"
}

func (c *FileCommand) Help() string {
	helpText := `
使用: vault file <子命令> [选项] [参数]

  此命令包含用于与安全库的键值存储区交互的子命令。
  下面是一些简单的示例，更详细的示例可以在子命令或文档中找到。

  加密文件，并输出文件：

      $ vault file encrypt -password=密码 -input=文件地址 -output=文件地址

  解密文件，并输出文件：

      $ vault file decrypt -password=密码 -input=文件地址 -output=文件地址

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *FileCommand) Run(args []string) int {
	return cli.RunResultHelp
}
