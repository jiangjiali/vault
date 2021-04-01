package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVMetadataCommand)(nil)

type KVMetadataCommand struct {
	*BaseCommand
}

func (c *KVMetadataCommand) Synopsis() string {
	return "与键值存储元数据交互"
}

func (c *KVMetadataCommand) Help() string {
	helpText := `
是哦用: vault kv metadata <子命令> [选项] [参数]

  此命令包含用于与安全库的键值存储区中的元数据终结点交互的子命令。
  下面是一些简单的示例，更详细的示例可以在子命令或文档中找到。

  为键创建或更新元数据项：

      $ vault kv metadata put -max-versions=5 secret/foo

  获取密钥的元数据，这将提供有关每个现有版本的信息：

      $ vault kv metadata get secret/foo

  删除密钥和所有现有版本：

      $ vault kv metadata delete secret/foo

  有关详细的用法信息，请参阅各个子命令帮助。
`

	return strings.TrimSpace(helpText)
}

func (c *KVMetadataCommand) Run(args []string) int {
	return cli.RunResultHelp
}
