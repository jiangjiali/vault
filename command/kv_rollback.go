package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*KVRollbackCommand)(nil)
var _ cli.CommandAutocomplete = (*KVRollbackCommand)(nil)

type KVRollbackCommand struct {
	*BaseCommand

	flagVersion int
}

func (c *KVRollbackCommand) Synopsis() string {
	return "回滚到以前版本的数据"
}

func (c *KVRollbackCommand) Help() string {
	helpText := `
使用: vault kv rollback [选项] KEY

  *注意*：这只支持KV v2机密引擎。

  将给定的先前版本还原为给定路径下的当前版本。该值将作为新版本写入；
  例如，如果当前版本为5，回滚版本为2，则来自版本2的数据将变为版本6。

      $ vault kv rollback -version=2 secret/foo

  下面详细介绍了其他标志和更高级的用例。

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *KVRollbackCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	// Common Options
	f := set.NewFlagSet("命令选项")

	f.IntVar(&IntVar{
		Name:   "version",
		Target: &c.flagVersion,
		Usage:  `指定应再次成为当前版本的版本号。`,
	})

	return set
}

func (c *KVRollbackCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *KVRollbackCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *KVRollbackCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	var version *int
	f.Visit(func(fl *flag.Flag) {
		if fl.Name == "version" {
			version = &c.flagVersion
		}
	})

	args = f.Args()

	switch {
	case len(args) != 1:
		c.UI.Error(fmt.Sprintf("Invalid number of arguments (expected 1, got %d)", len(args)))
		return 1
	case version == nil:
		c.UI.Error(fmt.Sprintf("Version flag must be specified"))
		return 1
	case c.flagVersion <= 0:
		c.UI.Error(fmt.Sprintf("Invalid value %d for the version flag", c.flagVersion))
		return 1
	}

	var err error
	path := sanitizePath(args[0])

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	mountPath, v2, err := isKVv2(path, client)
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	if !v2 {
		c.UI.Error(fmt.Sprintf("K/V engine mount must be version 2 for rollback support"))
		return 2
	}

	path = addPrefixToVKVPath(path, mountPath, "data")

	// First, do a read to get the current version for check-and-set
	var meta map[string]interface{}
	{
		secret, err := kvReadRequest(client, path, nil)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error doing pre-read at %s: %s", path, err))
			return 2
		}

		// Make sure a value already exists
		if secret == nil || secret.Data == nil {
			c.UI.Error(fmt.Sprintf("No value found at %s", path))
			return 2
		}

		// Verify metadata found
		rawMeta, ok := secret.Data["metadata"]
		if !ok || rawMeta == nil {
			c.UI.Error(fmt.Sprintf("No metadata found at %s; rollback only works on existing data", path))
			return 2
		}
		meta, ok = rawMeta.(map[string]interface{})
		if !ok {
			c.UI.Error(fmt.Sprintf("Metadata found at %s is not the expected type (JSON object)", path))
			return 2
		}
		if meta == nil {
			c.UI.Error(fmt.Sprintf("No metadata found at %s; rollback only works on existing data", path))
			return 2
		}
	}

	casVersion := meta["version"]

	// Set the version parameter
	versionParam := map[string]string{
		"version": fmt.Sprintf("%d", c.flagVersion),
	}

	// Now run it again and read the version we want to roll back to
	var data map[string]interface{}
	{
		secret, err := kvReadRequest(client, path, versionParam)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error doing pre-read at %s: %s", path, err))
			return 2
		}

		// Make sure a value already exists
		if secret == nil || secret.Data == nil {
			c.UI.Error(fmt.Sprintf("No value found at %s", path))
			return 2
		}

		// Verify metadata found
		rawMeta, ok := secret.Data["metadata"]
		if !ok || rawMeta == nil {
			c.UI.Error(fmt.Sprintf("No metadata found at %s; rollback only works on existing data", path))
			return 2
		}
		meta, ok := rawMeta.(map[string]interface{})
		if !ok {
			c.UI.Error(fmt.Sprintf("Metadata found at %s is not the expected type (JSON object)", path))
			return 2
		}
		if meta == nil {
			c.UI.Error(fmt.Sprintf("No metadata found at %s; rollback only works on existing data", path))
			return 2
		}

		// Verify it hasn't been deleted
		if meta["deletion_time"] != nil && meta["deletion_time"].(string) != "" {
			c.UI.Error(fmt.Sprintf("Cannot roll back to a version that has been deleted"))
			return 2
		}

		if meta["destroyed"] != nil && meta["destroyed"].(bool) {
			c.UI.Error(fmt.Sprintf("Cannot roll back to a version that has been destroyed"))
			return 2
		}

		// Verify old data found
		rawData, ok := secret.Data["data"]
		if !ok || rawData == nil {
			c.UI.Error(fmt.Sprintf("No data found at %s; rollback only works on existing data", path))
			return 2
		}
		data, ok = rawData.(map[string]interface{})
		if !ok {
			c.UI.Error(fmt.Sprintf("Data found at %s is not the expected type (JSON object)", path))
			return 2
		}
		if data == nil {
			c.UI.Error(fmt.Sprintf("No data found at %s; rollback only works on existing data", path))
			return 2
		}
	}

	secret, err := client.Logical().Write(path, map[string]interface{}{
		"data": data,
		"options": map[string]interface{}{
			"cas": casVersion,
		},
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error writing data to %s: %s", path, err))
		return 2
	}
	if secret == nil {
		// Don't output anything unless using the "table" format
		if Format(c.UI) == "table" {
			c.UI.Info(fmt.Sprintf("Success! Data written to: %s", path))
		}
		return 0
	}

	if c.flagField != "" {
		return PrintRawField(c.UI, secret, c.flagField)
	}

	return OutputSecret(c.UI, secret)
}
