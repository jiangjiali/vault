package command

import (
	"flag"
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strconv"
	"strings"
	"time"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/consts"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*AuthEnableCommand)(nil)
var _ cli.CommandAutocomplete = (*AuthEnableCommand)(nil)

type AuthEnableCommand struct {
	*BaseCommand

	flagDescription               string
	flagPath                      string
	flagDefaultLeaseTTL           time.Duration
	flagMaxLeaseTTL               time.Duration
	flagAuditNonHMACRequestKeys   []string
	flagAuditNonHMACResponseKeys  []string
	flagListingVisibility         string
	flagPluginName                string
	flagPassthroughRequestHeaders []string
	flagAllowedResponseHeaders    []string
	flagOptions                   map[string]string
	flagLocal                     bool
	flagSealWrap                  bool
	flagTokenType                 string
	flagVersion                   int
}

func (c *AuthEnableCommand) Synopsis() string {
	return "启用新的身份验证方法"
}

func (c *AuthEnableCommand) Help() string {
	helpText := `
使用: vault auth enable [选项] TYPE

  启用新的身份验证方法。
  身份验证方法负责对用户或计算机进行身份验证，
  并为其分配策略，以便用户或计算机访问安全库。

  在 userpass/ 处启用 userpass 认证方法：

      $ vault auth enable userpass

  在 auth-prod/ 处启用 userpass 认证方法：

      $ vault auth enable -path=auth-prod userpass

  启用自定义身份验证插件（在插件注册表中注册后）：

      $ vault auth enable -path=my-auth -plugin-name=my-auth-plugin plugin

      或（首选方式）：

      $ vault auth enable -path=my-auth my-auth-plugin

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *AuthEnableCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP)

	f := set.NewFlagSet("命令选项")

	f.StringVar(&StringVar{
		Name:       "description",
		Target:     &c.flagDescription,
		Completion: complete.PredictAnything,
		Usage:      "用于此身份验证方法的人性化描述。",
	})

	f.StringVar(&StringVar{
		Name:       "path",
		Target:     &c.flagPath,
		Default:    "", // The default is complex, so we have to manually document
		Completion: complete.PredictAnything,
		Usage:      "验证方法可访问的位置。这在所有身份验证方法中必须是唯一的。这默认为认证方法的“type”。认证方法可在“/auth/<path>”处访问。",
	})

	f.DurationVar(&DurationVar{
		Name:       "default-lease-ttl",
		Target:     &c.flagDefaultLeaseTTL,
		Completion: complete.PredictAnything,
		Usage:      "此身份验证方法的默认租约TTL。如果未指定，则默认为安全库服务器的全局配置的默认租约TTL。",
	})

	f.DurationVar(&DurationVar{
		Name:       "max-lease-ttl",
		Target:     &c.flagMaxLeaseTTL,
		Completion: complete.PredictAnything,
		Usage:      "此身份验证方法的最大租约TTL。如果未指定，则默认为Vault服务器的全局配置的最大租约TTL。",
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

	f.StringVar(&StringVar{
		Name:       "plugin-name",
		Target:     &c.flagPluginName,
		Completion: c.PredictVaultPlugins(consts.PluginTypeCredential),
		Usage:      "验证方法插件的名称。此插件名称必须已存在于Vault服务器的插件目录中。",
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
		Usage:   "将认证方法标记为local only。本地身份验证方法不会被复制，也不会被复制删除。",
	})

	f.BoolVar(&BoolVar{
		Name:    "seal-wrap",
		Target:  &c.flagSealWrap,
		Default: false,
		Usage:   "在机密引擎中启用临界值的密封包装。",
	})

	f.StringVar(&StringVar{
		Name:   flagNameTokenType,
		Target: &c.flagTokenType,
		Usage:  "设置装载的强制令牌类型。",
	})

	f.IntVar(&IntVar{
		Name:    "version",
		Target:  &c.flagVersion,
		Default: 0,
		Usage:   "选择要运行的身份验证方法的版本。并非所有身份验证方法都支持。",
	})

	return set
}

func (c *AuthEnableCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultAvailableAuths()
}

func (c *AuthEnableCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *AuthEnableCommand) Run(args []string) int {
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

	authType := strings.TrimSpace(args[0])
	if authType == "plugin" {
		authType = c.flagPluginName
	}

	// If no path is specified, we default the path to the backend type
	// or use the plugin name if it's a plugin backend
	authPath := c.flagPath
	if authPath == "" {
		if authType == "plugin" {
			authPath = c.flagPluginName
		} else {
			authPath = authType
		}
	}

	// Append a trailing slash to indicate it's a path in output
	authPath = ensureTrailingSlash(authPath)

	if c.flagVersion > 0 {
		if c.flagOptions == nil {
			c.flagOptions = make(map[string]string)
		}
		c.flagOptions["version"] = strconv.Itoa(c.flagVersion)
	}

	authOpts := &api.EnableAuthOptions{
		Type:        authType,
		Description: c.flagDescription,
		Local:       c.flagLocal,
		SealWrap:    c.flagSealWrap,
		Config: api.AuthConfigInput{
			DefaultLeaseTTL: c.flagDefaultLeaseTTL.String(),
			MaxLeaseTTL:     c.flagMaxLeaseTTL.String(),
		},
		Options: c.flagOptions,
	}

	// Set these values only if they are provided in the CLI
	f.Visit(func(fl *flag.Flag) {
		if fl.Name == flagNameAuditNonHMACRequestKeys {
			authOpts.Config.AuditNonHMACRequestKeys = c.flagAuditNonHMACRequestKeys
		}

		if fl.Name == flagNameAuditNonHMACResponseKeys {
			authOpts.Config.AuditNonHMACResponseKeys = c.flagAuditNonHMACResponseKeys
		}

		if fl.Name == flagNameListingVisibility {
			authOpts.Config.ListingVisibility = c.flagListingVisibility
		}

		if fl.Name == flagNamePassthroughRequestHeaders {
			authOpts.Config.PassthroughRequestHeaders = c.flagPassthroughRequestHeaders
		}

		if fl.Name == flagNameAllowedResponseHeaders {
			authOpts.Config.AllowedResponseHeaders = c.flagAllowedResponseHeaders
		}

		if fl.Name == flagNameTokenType {
			authOpts.Config.TokenType = c.flagTokenType
		}
	})

	if err := client.Sys().EnableAuthWithOptions(authPath, authOpts); err != nil {
		c.UI.Error(fmt.Sprintf("Error enabling %s auth: %s", authType, err))
		return 2
	}

	authThing := authType + " auth method"
	if authType == "plugin" {
		authThing = c.flagPluginName + " plugin"
	}
	c.UI.Output(fmt.Sprintf("Success! Enabled %s at: %s", authThing, authPath))
	return 0
}
