package command

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jiangjiali/vault/api"
	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*SecretsTuneCommand)(nil)
var _ cli.CommandAutocomplete = (*SecretsTuneCommand)(nil)

type SecretsTuneCommand struct {
	*BaseCommand

	flagAuditNonHMACRequestKeys  []string
	flagAuditNonHMACResponseKeys []string
	flagDefaultLeaseTTL          time.Duration
	flagDescription              string
	flagListingVisibility        string
	flagMaxLeaseTTL              time.Duration
	flagOptions                  map[string]string
	flagVersion                  int
}

func (c *SecretsTuneCommand) Synopsis() string {
	return "调整机密引擎配置"
}

func (c *SecretsTuneCommand) Help() string {
	helpText := `
使用: vault secrets tune [选项] PATH

  在给定PATH上调整机密引擎的配置选项。
  参数对应于启用机密引擎的PATH，而不是TYPE！

  优化PKI机密引擎的默认租约：

      $ vault secrets tune -default-lease-ttl=72h pki/

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *SecretsTuneCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP)

	f := set.NewFlagSet("命令选项")

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

	f.DurationVar(&DurationVar{
		Name:       "default-lease-ttl",
		Target:     &c.flagDefaultLeaseTTL,
		Default:    0,
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "此机密引擎的默认租约TTL。如果未指定，则默认为安全库服务器的全局配置的默认租约TTL，或先前为机密引擎配置的值。",
	})

	f.StringVar(&StringVar{
		Name:   flagNameDescription,
		Target: &c.flagDescription,
		Usage:  "对这个秘密引擎的人性化描述。这将覆盖当前存储的值（如果有）。",
	})

	f.StringVar(&StringVar{
		Name:   flagNameListingVisibility,
		Target: &c.flagListingVisibility,
		Usage:  "确定装载在特定于UI的列表终结点中的可见性。",
	})

	f.DurationVar(&DurationVar{
		Name:       "max-lease-ttl",
		Target:     &c.flagMaxLeaseTTL,
		Default:    0,
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "此机密引擎的最大租约TTL。如果未指定，则默认为安全库服务器的全局配置的最大租约TTL，或先前为机密引擎配置的值。",
	})

	f.StringMapVar(&StringMapVar{
		Name:       "options",
		Target:     &c.flagOptions,
		Completion: complete.PredictAnything,
		Usage:      "作为安装选项的Key=value提供的Key-value对。可以多次指定。",
	})

	f.IntVar(&IntVar{
		Name:    "version",
		Target:  &c.flagVersion,
		Default: 0,
		Usage:   "选择要运行的引擎版本。并非所有机密引擎都支持。",
	})

	return set
}

func (c *SecretsTuneCommand) AutocompleteArgs() complete.Predictor {
	return c.PredictVaultMounts()
}

func (c *SecretsTuneCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *SecretsTuneCommand) Run(args []string) int {
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

	if c.flagVersion > 0 {
		if c.flagOptions == nil {
			c.flagOptions = make(map[string]string)
		}
		c.flagOptions["version"] = strconv.Itoa(c.flagVersion)
	}

	// Append a trailing slash to indicate it's a path in output
	mountPath := ensureTrailingSlash(sanitizePath(args[0]))

	mountConfigInput := api.MountConfigInput{
		DefaultLeaseTTL: ttlToAPI(c.flagDefaultLeaseTTL),
		MaxLeaseTTL:     ttlToAPI(c.flagMaxLeaseTTL),
		Options:         c.flagOptions,
	}

	// Set these values only if they are provided in the CLI
	f.Visit(func(fl *flag.Flag) {
		if fl.Name == flagNameAuditNonHMACRequestKeys {
			mountConfigInput.AuditNonHMACRequestKeys = c.flagAuditNonHMACRequestKeys
		}

		if fl.Name == flagNameAuditNonHMACResponseKeys {
			mountConfigInput.AuditNonHMACResponseKeys = c.flagAuditNonHMACResponseKeys
		}

		if fl.Name == flagNameDescription {
			mountConfigInput.Description = &c.flagDescription
		}

		if fl.Name == flagNameListingVisibility {
			mountConfigInput.ListingVisibility = c.flagListingVisibility
		}
	})

	if err := client.Sys().TuneMount(mountPath, mountConfigInput); err != nil {
		c.UI.Error(fmt.Sprintf("Error tuning secrets engine %s: %s", mountPath, err))
		return 2
	}

	c.UI.Output(fmt.Sprintf("Success! Tuned the secrets engine at: %s", mountPath))
	return 0
}
