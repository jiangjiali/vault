package command

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jiangjiali/vault/api"
	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/consts"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*SecretsEnableCommand)(nil)
var _ cli.CommandAutocomplete = (*SecretsEnableCommand)(nil)

type SecretsEnableCommand struct {
	*BaseCommand

	flagDescription               string
	flagPath                      string
	flagDefaultLeaseTTL           time.Duration
	flagMaxLeaseTTL               time.Duration
	flagAuditNonHMACRequestKeys   []string
	flagAuditNonHMACResponseKeys  []string
	flagListingVisibility         string
	flagPassthroughRequestHeaders []string
	flagAllowedResponseHeaders    []string
	flagForceNoCache              bool
	flagPluginName                string
	flagOptions                   map[string]string
	flagLocal                     bool
	flagSealWrap                  bool
	flagVersion                   int
}

func (c *SecretsEnableCommand) Synopsis() string {
	return "启用机密引擎"
}

func (c *SecretsEnableCommand) Help() string {
	helpText := `
使用: vault secrets enable [选项] TYPE

  启用秘密引擎。默认情况下，机密引擎在其TYPE对应的路径上启用，
  但用户可以使用 -path 选项自定义路径。

  一旦启用，安全库将把所有以路径开头的请求路由到机密引擎。

  在 kv/ 下启动秘密引擎：

      $ vault secrets enable kv

  启用自定义插件（在插件注册表中注册后）：

      $ vault secrets enable -path=my-secrets -plugin-name=my-plugin plugin

  或（首选方式）：

      $ vault secrets enable -path=my-secrets my-plugin

  有关机密引擎和示例的完整列表，请参阅文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *SecretsEnableCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP)

	f := set.NewFlagSet("命令选项")

	f.StringVar(&StringVar{
		Name:       "description",
		Target:     &c.flagDescription,
		Completion: complete.PredictAnything,
		Usage:      "此引擎的人性化描述。",
	})

	f.StringVar(&StringVar{
		Name:       "path",
		Target:     &c.flagPath,
		Default:    "", // The default is complex, so we have to manually document
		Completion: complete.PredictAnything,
		Usage:      "秘密引擎可以进入的地方。这一定是独一无二的跨所有秘密引擎。这默认为机密引擎的“类型”。",
	})

	f.DurationVar(&DurationVar{
		Name:       "default-lease-ttl",
		Target:     &c.flagDefaultLeaseTTL,
		Completion: complete.PredictAnything,
		Usage:      "此机密引擎的默认租约TTL。如果未指定，则默认为安全库服务器的全局配置的默认租约TTL。",
	})

	f.DurationVar(&DurationVar{
		Name:       "max-lease-ttl",
		Target:     &c.flagMaxLeaseTTL,
		Completion: complete.PredictAnything,
		Usage:      "此机密引擎的最大租约TTL。如果未指定，则默认为Vault服务器的全局配置的最大租约TTL。",
	})

	f.StringSliceVar(&StringSliceVar{
		Name:   flagNameAuditNonHMACRequestKeys,
		Target: &c.flagAuditNonHMACRequestKeys,
		Usage:  "逗号分隔的字符串或请求数据对象中的审核设备不会HMAC的键列表。",
	})

	f.StringSliceVar(&StringSliceVar{
		Name:   flagNameAuditNonHMACResponseKeys,
		Target: &c.flagAuditNonHMACResponseKeys,
		Usage:  "逗号分隔的字符串或键列表，这些键不会被响应数据对象中的审核设备HMAC。",
	})

	f.StringVar(&StringVar{
		Name:   flagNameListingVisibility,
		Target: &c.flagListingVisibility,
		Usage:  "确定装载在特定于UI的列表终结点中的可见性。",
	})

	f.StringSliceVar(&StringSliceVar{
		Name:   flagNamePassthroughRequestHeaders,
		Target: &c.flagPassthroughRequestHeaders,
		Usage:  "逗号分隔字符串或将发送到插件的请求头值列表。",
	})

	f.StringSliceVar(&StringSliceVar{
		Name:   flagNameAllowedResponseHeaders,
		Target: &c.flagAllowedResponseHeaders,
		Usage:  "逗号分隔字符串或允许插件设置的响应头值列表。",
	})

	f.BoolVar(&BoolVar{
		Name:    "force-no-cache",
		Target:  &c.flagForceNoCache,
		Default: false,
		Usage:   "强制机密引擎禁用缓存。如果未指定，则默认为安全库服务器的全局配置的缓存设置。这不会影响底层加密数据存储的缓存。",
	})

	f.StringVar(&StringVar{
		Name:       "plugin-name",
		Target:     &c.flagPluginName,
		Completion: c.PredictVaultPlugins(consts.PluginTypeSecrets),
		Usage:      "引擎的秘密名称。此插件名称必须已存在于安全库的插件目录中。",
	})

	f.StringMapVar(&StringMapVar{
		Name:       "options",
		Target:     &c.flagOptions,
		Completion: complete.PredictAnything,
		Usage:      "作为安装选项的Key=value提供的Key-value对。可以多次指定。",
	})

	f.BoolVar(&BoolVar{
		Name:    "local",
		Target:  &c.flagLocal,
		Default: false,
		Usage:   "将机密引擎标记为仅本地。本地引擎不会被复制或删除。",
	})

	f.BoolVar(&BoolVar{
		Name:    "seal-wrap",
		Target:  &c.flagSealWrap,
		Default: false,
		Usage:   "在机密引擎中启用临界值的密封包装。",
	})

	f.IntVar(&IntVar{
		Name:    "version",
		Target:  &c.flagVersion,
		Default: 0,
		Usage:   "选择要运行的引擎版本。并非所有机密引擎都支持。",
	})

	return set
}

func (c *SecretsEnableCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultAvailableMounts()
}

func (c *SecretsEnableCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *SecretsEnableCommand) Run(args []string) int {
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

	// Get the engine type type (first arg)
	engineType := strings.TrimSpace(args[0])
	if engineType == "plugin" {
		engineType = c.flagPluginName
	}

	// If no path is specified, we default the path to the backend type
	// or use the plugin name if it's a plugin backend
	mountPath := c.flagPath
	if mountPath == "" {
		if engineType == "plugin" {
			mountPath = c.flagPluginName
		} else {
			mountPath = engineType
		}
	}

	if c.flagVersion > 0 {
		if c.flagOptions == nil {
			c.flagOptions = make(map[string]string)
		}
		c.flagOptions["version"] = strconv.Itoa(c.flagVersion)
	}

	// Append a trailing slash to indicate it's a path in output
	mountPath = ensureTrailingSlash(mountPath)

	// Build mount input
	mountInput := &api.MountInput{
		Type:        engineType,
		Description: c.flagDescription,
		Local:       c.flagLocal,
		SealWrap:    c.flagSealWrap,
		Config: api.MountConfigInput{
			DefaultLeaseTTL: c.flagDefaultLeaseTTL.String(),
			MaxLeaseTTL:     c.flagMaxLeaseTTL.String(),
			ForceNoCache:    c.flagForceNoCache,
		},
		Options: c.flagOptions,
	}

	// Set these values only if they are provided in the CLI
	f.Visit(func(fl *flag.Flag) {
		if fl.Name == flagNameAuditNonHMACRequestKeys {
			mountInput.Config.AuditNonHMACRequestKeys = c.flagAuditNonHMACRequestKeys
		}

		if fl.Name == flagNameAuditNonHMACResponseKeys {
			mountInput.Config.AuditNonHMACResponseKeys = c.flagAuditNonHMACResponseKeys
		}

		if fl.Name == flagNameListingVisibility {
			mountInput.Config.ListingVisibility = c.flagListingVisibility
		}

		if fl.Name == flagNamePassthroughRequestHeaders {
			mountInput.Config.PassthroughRequestHeaders = c.flagPassthroughRequestHeaders
		}

		if fl.Name == flagNameAllowedResponseHeaders {
			mountInput.Config.AllowedResponseHeaders = c.flagAllowedResponseHeaders
		}
	})

	if err := client.Sys().Mount(mountPath, mountInput); err != nil {
		c.UI.Error(fmt.Sprintf("Error enabling: %s", err))
		return 2
	}

	thing := engineType + " secrets engine"
	if engineType == "plugin" {
		thing = c.flagPluginName + " plugin"
	}
	c.UI.Output(fmt.Sprintf("Success! Enabled the %s at: %s", thing, mountPath))
	return 0
}
