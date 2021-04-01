package command

import (
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*LeaseLookupCommand)(nil)
var _ cli.CommandAutocomplete = (*LeaseLookupCommand)(nil)

type LeaseLookupCommand struct {
	*BaseCommand
}

func (c *LeaseLookupCommand) Synopsis() string {
	return "查寻机密租约"
}

func (c *LeaseLookupCommand) Help() string {
	helpText := `
使用: vault lease lookup ID

  查找机密的租约信息。

  vault里的每一个机密都有一个租约。用户可以通过引用租约ID来查找有关租约的信息。

  查找机密租约:

      $ vault lease lookup database/creds/readonly/2f6a614c...

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *LeaseLookupCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	return set
}

func (c *LeaseLookupCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictAnything
}

func (c *LeaseLookupCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *LeaseLookupCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	leaseID := ""

	args = f.Args()
	switch len(args) {
	case 0:
		c.UI.Error("找不到ID!")
		return 1
	case 1:
		leaseID = strings.TrimSpace(args[0])
	default:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	secret, err := client.Sys().Lookup(leaseID)
	if err != nil {
		c.UI.Error(fmt.Sprintf("error looking up lease id %s: %s", leaseID, err))
		return 2
	}

	return OutputSecret(c.UI, secret)
}
