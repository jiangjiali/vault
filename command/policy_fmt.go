package command

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/hclutil/hcl/hcl/printer"
	"github.com/jiangjiali/vault/sdk/helper/homedir"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/helper/namespace"
	"github.com/jiangjiali/vault/vault"
)

var _ cli.Command = (*PolicyFmtCommand)(nil)
var _ cli.CommandAutocomplete = (*PolicyFmtCommand)(nil)

type PolicyFmtCommand struct {
	*BaseCommand
}

func (c *PolicyFmtCommand) Synopsis() string {
	return "在磁盘上格式化策略"
}

func (c *PolicyFmtCommand) Help() string {
	helpText := `
使用: vault policy fmt [选项] PATH

  将本地策略文件格式化为策略规范。
  此命令将用正确格式化的策略文件内容覆盖给定路径的文件。

  格式化本地文件“my-policy.hcl“作为策略文件：

      $ vault policy fmt my-policy.hcl

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *PolicyFmtCommand) Flags() *FlagSets {
	return c.flagSet(FlagSetNone)
}

func (c *PolicyFmtCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictFiles("*.hcl")
}

func (c *PolicyFmtCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *PolicyFmtCommand) Run(args []string) int {
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

	// Get the filepath, accounting for ~ and stuff
	path, err := homedir.Expand(strings.TrimSpace(args[0]))
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to expand path: %s", err))
		return 1
	}

	// Read the entire contents into memory - it would be nice if we could use
	// a buffer, but hcl wants the full contents.
	b, err := ioutil.ReadFile(path)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading source file: %s", err))
		return 1
	}

	// Actually parse the policy. We always use the root namespace here because
	// we don't want to modify the results.
	if _, err := vault.ParseACLPolicy(namespace.RootNamespace, string(b)); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	// Generate final contents
	result, err := printer.Format(b)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error printing result: %s", err))
		return 1
	}

	// Write them back out
	if err := ioutil.WriteFile(path, result, 0644); err != nil {
		c.UI.Error(fmt.Sprintf("Error writing result: %s", err))
		return 1
	}

	c.UI.Output(fmt.Sprintf("Success! Formatted policy: %s", path))
	return 0
}
