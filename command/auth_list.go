package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"sort"
	"strconv"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*AuthListCommand)(nil)
var _ cli.CommandAutocomplete = (*AuthListCommand)(nil)

type AuthListCommand struct {
	*BaseCommand

	flagDetailed bool
}

func (c *AuthListCommand) Synopsis() string {
	return "列出启用的身份验证方法"
}

func (c *AuthListCommand) Help() string {
	helpText := `
使用: vault auth list [选项]

  列出Vault服务器上启用的验证方法。
  此命令还输出有关方法的信息，包括配置和人性化描述。
  TTL为“system”表示系统默认值正在使用中。

  列出所有启用的身份验证方法：

      $ vault auth list

  列出所有启用的身份验证方法及其详细输出：

      $ vault auth list -detailed

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *AuthListCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:    "detailed",
		Target:  &c.flagDetailed,
		Default: false,
		Usage: "Print detailed information such as configuration and replication " +
			"status about each auth method. This option is only applicable to " +
			"table-formatted output.",
	})

	return set
}

func (c *AuthListCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *AuthListCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *AuthListCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	if len(args) > 0 {
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 0, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	auths, err := client.Sys().ListAuth()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error listing enabled authentications: %s", err))
		return 2
	}

	switch Format(c.UI) {
	case "table":
		if c.flagDetailed {
			c.UI.Output(tableOutput(c.detailedMounts(auths), nil))
			return 0
		}
		c.UI.Output(tableOutput(c.simpleMounts(auths), nil))
		return 0
	default:
		return OutputData(c.UI, auths)
	}
}

func (c *AuthListCommand) simpleMounts(auths map[string]*api.AuthMount) []string {
	paths := make([]string, 0, len(auths))
	for path := range auths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	out := []string{"Path | Type | Accessor | Description"}
	for _, path := range paths {
		mount := auths[path]
		out = append(out, fmt.Sprintf("%s | %s | %s | %s", path, mount.Type, mount.Accessor, mount.Description))
	}

	return out
}

func (c *AuthListCommand) detailedMounts(auths map[string]*api.AuthMount) []string {
	paths := make([]string, 0, len(auths))
	for path := range auths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	calcTTL := func(typ string, ttl int) string {
		switch {
		case typ == "system", typ == "public":
			return ""
		case ttl != 0:
			return strconv.Itoa(ttl)
		default:
			return "system"
		}
	}

	out := []string{"Path | Plugin | Accessor | Default TTL | Max TTL | Token Type | Replication | Seal Wrap | Options | Description | UUID"}
	for _, path := range paths {
		mount := auths[path]

		defaultTTL := calcTTL(mount.Type, mount.Config.DefaultLeaseTTL)
		maxTTL := calcTTL(mount.Type, mount.Config.MaxLeaseTTL)

		replication := "replicated"
		if mount.Local {
			replication = "local"
		}

		pluginName := mount.Type
		if pluginName == "plugin" {
			pluginName = mount.Config.PluginName
		}

		out = append(out, fmt.Sprintf("%s | %s | %s | %s | %s | %s | %s | %t | %v | %s | %s",
			path,
			pluginName,
			mount.Accessor,
			defaultTTL,
			maxTTL,
			mount.Config.TokenType,
			replication,
			mount.SealWrap,
			mount.Options,
			mount.Description,
			mount.UUID,
		))
	}

	return out
}
