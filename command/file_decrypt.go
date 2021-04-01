package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/helper/zcrypto"
)

var _ cli.Command = (*FileDecryptCommand)(nil)
var _ cli.CommandAutocomplete = (*FileDecryptCommand)(nil)

type FileDecryptCommand struct {
	*BaseCommand

	flagPassword   string
	flagInputFile  string
	flagOutputFile string
}

func (c *FileDecryptCommand) Synopsis() string {
	return "文件解密"
}

func (c *FileDecryptCommand) Help() string {
	helpText := `

使用: vault file decrypt [选项]

  给指定文件解密。

  解密inputFile文件，并输出文件：

      $ vault file decrypt -password=密码 -input=文件地址 -output=文件地址

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *FileDecryptCommand) Flags() *FlagSets {
	set := NewFlagSets()

	f := set.NewFlagSet("选项")

	f.StringVar(&StringVar{
		Name:       "password",
		Target:     &c.flagPassword,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "解密文件的密码",
	})

	f.StringVar(&StringVar{
		Name:       "input",
		Target:     &c.flagInputFile,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "输出文件地址 ",
	})

	f.StringVar(&StringVar{
		Name:       "output",
		Target:     &c.flagOutputFile,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "输出文件地址",
	})

	return set
}

func (c *FileDecryptCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFolders()
}

func (c *FileDecryptCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *FileDecryptCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	if len(args) > 0 {
		c.UI.Error(fmt.Sprintf("参数太多（应为 0 个，获得了 %d 个）", len(args)))
		return 1
	}

	c.UI.Output("文件解密中...")

	err := zcrypto.CfbDecryptoFromFile(c.flagInputFile, c.flagOutputFile, c.flagPassword)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	c.UI.Output("解密完成")
	return 0
}
