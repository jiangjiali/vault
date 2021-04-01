package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVEnableVersioningCommand)(nil)
var _ cli.CommandAutocomplete = (*KVEnableVersioningCommand)(nil)

type KVEnableVersioningCommand struct {
	*BaseCommand
}

func (c *KVEnableVersioningCommand) Synopsis() string {
	return "开启KV存储的版本控制"
}

func (c *KVEnableVersioningCommand) Help() string {
	helpText := `
使用: vault kv enable-versions [选项] KEY

  此命令在提供的路径上打开后端的版本控制。

      $ vault kv enable-versions secret

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *KVEnableVersioningCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	return set
}

func (c *KVEnableVersioningCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *KVEnableVersioningCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVEnableVersioningCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	switch {
	case len(args) < 1:
		c.UI.Error(fmt.Sprintf("Not enough arguments (expected 1, got %d)", len(args)))
		return 1
	case len(args) > 1:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	// Append a trailing slash to indicate it's a path in output
	mountPath := ensureTrailingSlash(sanitizePath(args[0]))

	if err := client.Sys().TuneMount(mountPath, api.MountConfigInput{
		Options: map[string]string{
			"version": "2",
		},
	}); err != nil {
		c.UI.Error(fmt.Sprintf("Error tuning secrets engine %s: %s", mountPath, err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Tuned the secrets engine at: %s", mountPath))
	return 0
}
