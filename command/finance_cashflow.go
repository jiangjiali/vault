package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/helper/zcrypto"
)

var _ cli.Command = (*FinanceCashflowCommand)(nil)
var _ cli.CommandAutocomplete = (*FinanceCashflowCommand)(nil)

type FinanceCashflowCommand struct {
	*BaseCommand

	flagPassword   string
	flagInputFile  string
	flagOutputFile string
}

func (c *FinanceCashflowCommand) Synopsis() string {
	return "文件加密"
}

func (c *FinanceCashflowCommand) Help() string {
	helpText := `

使用: vault file encrypt [选项]

  给指定文件加密。

  加密inputFile文件，并输出文件：

      $ vault file encrypt -password=密码 -input=文件地址 -output=文件地址

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *FinanceCashflowCommand) Flags() *FlagSets {
	set := NewFlagSets()

	f := set.NewFlagSet("选项")

	f.StringVar(&StringVar{
		Name:       "password",
		Target:     &c.flagPassword,
		Default:    "123456",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "加密文件的密码",
	})

	f.StringVar(&StringVar{
		Name:       "input",
		Target:     &c.flagInputFile,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "输入文件地址",
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

func (c *FinanceCashflowCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultFolders()
}

func (c *FinanceCashflowCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *FinanceCashflowCommand) Run(args []string) int {
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

	c.UI.Output("文件加密中...")

	err := zcrypto.CfbEncryptoToFile(c.flagInputFile, c.flagOutputFile, c.flagPassword)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	c.UI.Output("加密完成")
	return 0
}
