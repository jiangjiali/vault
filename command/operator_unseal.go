package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"io"
	"os"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/helper/password"
)

var _ cli.Command = (*OperatorUnsealCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorUnsealCommand)(nil)

type OperatorUnsealCommand struct {
	*BaseCommand

	flagReset   bool
	flagMigrate bool

	testOutput io.Writer // for tests
}

func (c *OperatorUnsealCommand) Synopsis() string {
	return "解封服务器"
}

func (c *OperatorUnsealCommand) Help() string {
	helpText := `
使用: vault operator unseal [选项] [KEY]

  提供一部分主密钥以解封安全库服务器。保险库以密封状态开始。
  它在解封之前不能执行操作。此命令接受主密钥的一部分（“解封密钥”）。

  解封密钥可以作为命令的参数提供，但不建议这样做，
  因为解封密钥将在您的历史记录中可用：

      $ vault operator unseal IXyR0OJnSFobekZMMCKCoVEpT7wI6l+USMzE3IcyDyo=

  相反，运行不带参数的命令，它将提示输入键：

      $ vault operator unseal
      Key (will be hidden): IXyR0OJnSFobekZMMCKCoVEpT7wI6l+USMzE3IcyDyo=

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorUnsealCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:       "reset",
		Aliases:    []string{},
		Target:     &c.flagReset,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "放弃先前输入的所有密钥以进行解封过程。",
	})

	f.BoolVar(&BoolVar{
		Name:       "migrate",
		Aliases:    []string{},
		Target:     &c.flagMigrate,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "说明提供此共享是为了使其成为密封迁移过程的一部分。",
	})

	return set
}

func (c *OperatorUnsealCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictAnything
}

func (c *OperatorUnsealCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorUnsealCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	unsealKey := ""

	args = f.Args()
	switch len(args) {
	case 0:
		// We will prompt for the unseal key later
	case 1:
		unsealKey = strings.TrimSpace(args[0])
	default:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	if c.flagReset {
		status, err := client.Sys().ResetUnsealProcess()
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error resetting unseal process: %s", err))
			return 2
		}
		return OutputSealStatus(c.UI, client, status)
	}

	if unsealKey == "" {
		// Override the output
		writer := (io.Writer)(os.Stdout)
		if c.testOutput != nil {
			writer = c.testOutput
		}

		fmt.Fprintf(writer, "Unseal Key (will be hidden): ")
		value, err := password.Read(os.Stdin)
		fmt.Fprintf(writer, "\n")
		if err != nil {
			c.UI.Error(wrapAtLength(fmt.Sprintf("An error occurred attempting to "+
				"ask for an unseal key. The raw error message is shown below, but "+
				"usually this is because you attempted to pipe a value into the "+
				"unseal command or you are executing outside of a terminal (tty). "+
				"You should run the unseal command from a terminal for maximum "+
				"security. If this is not an option, the unseal key can be provided "+
				"as the first argument to the unseal command. The raw error "+
				"was:\n\n%s", err)))
			return 1
		}
		unsealKey = strings.TrimSpace(value)
	}

	status, err := client.Sys().UnsealWithOptions(&api.UnsealOpts{
		Key:     unsealKey,
		Migrate: c.flagMigrate,
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error unsealing: %s", err))
		return 2
	}

	return OutputSealStatus(c.UI, client, status)
}
