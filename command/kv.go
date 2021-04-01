package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVCommand)(nil)

type KVCommand struct {
	*BaseCommand
}

func (c *KVCommand) Synopsis() string {
	return "与键值存储交互"
}

func (c *KVCommand) Help() string {
	helpText := `
使用: vault kv <子命令> [选项] [参数]

  此命令包含用于与安全库的键值存储区交互的子命令。
  下面是一些简单的示例，更详细的示例可以在子命令或文档中找到。

  在“secret”挂载中创建或更新名为“foo”的密钥，值为“bar=baz”：

      $ vault kv put secret/foo bar=baz

  读回此值：

      $ vault kv get secret/foo

  获取密钥的元数据：

      $ vault kv metadata get secret/foo
	  
  获取密钥的特定版本：

      $ vault kv get -version=1 secret/foo

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *KVCommand) Run(args []string) int {
	return cli.RunResultHelp
}
