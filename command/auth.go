package command

import (
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*AuthCommand)(nil)

type AuthCommand struct {
	*BaseCommand
}

func (c *AuthCommand) Synopsis() string {
	return "与身份验证方法交互"
}

func (c *AuthCommand) Help() string {
	return strings.TrimSpace(`
使用: vault auth <子命令> [选项] [参数]

  此命令将用于与服务器的身份验证方法交互的子命令分组。
  用户可以列出、启用、禁用和获取不同身份验证方法的帮助。

  要以用户或计算机身份验证到服务器，请改用“Vault login”命令。
  此命令用于与身份验证方法本身交互，而不是对服务器进行身份验证。

  列出所有启用的身份验证方法：

      $ vault auth list

  启用新的认证方法"userpass""；

      $ vault auth enable userpass

  获取有关如何对特定身份验证方法进行身份验证的详细帮助信息：

      $ vault auth help userpass

  有关详细的用法信息，请参阅各个子命令帮助。
`)
}

func (c *AuthCommand) Run(args []string) int {
	return cli.RunResultHelp
}
