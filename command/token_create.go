package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"
	"time"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
)

var _ cli.Command = (*TokenCreateCommand)(nil)
var _ cli.CommandAutocomplete = (*TokenCreateCommand)(nil)

type TokenCreateCommand struct {
	*BaseCommand

	flagID              string
	flagDisplayName     string
	flagTTL             time.Duration
	flagExplicitMaxTTL  time.Duration
	flagPeriod          time.Duration
	flagRenewable       bool
	flagOrphan          bool
	flagNoDefaultPolicy bool
	flagUseLimit        int
	flagRole            string
	flagType            string
	flagMetadata        map[string]string
	flagPolicies        []string
}

func (c *TokenCreateCommand) Synopsis() string {
	return "创建新令牌"
}

func (c *TokenCreateCommand) Help() string {
	helpText := `
使用: vault token create [选项]

  创建可用于身份验证的新令牌。此令牌将创建为当前已验证令牌的子令牌。
  生成的令牌将继承当前已验证令牌的所有策略和权限，
  除非显式定义要分配给令牌的子集列表策略。

  生存时间(ttl)也可以与令牌相关联。如果生存时间未与令牌关联，则无法续订它。
  如果生存时间与令牌关联，则除非续订，否则它将在该时间段后过期。

  使用令牌时，与令牌关联的元数据（用“-metadata”指定）将写入审核日志。

  如果指定了角色，则该角色可以重写此处指定的参数。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *TokenCreateCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)

	f := set.NewFlagSet("命令选项")

	f.StringVar(&StringVar{
		Name:       "id",
		Target:     &c.flagID,
		Completion: complete.PredictAnything,
		Usage:      "令牌的值。默认情况下，这是一个自动生成的36个字符的UUID。指定此值需要sudo权限。",
	})

	f.StringVar(&StringVar{
		Name:       "display-name",
		Target:     &c.flagDisplayName,
		Completion: complete.PredictAnything,
		Usage:      "与此令牌关联的名称。这是一个非敏感值，可用于帮助识别已创建的机密（例如：prefixes）。",
	})

	f.DurationVar(&DurationVar{
		Name:       "ttl",
		Target:     &c.flagTTL,
		Completion: complete.PredictAnything,
		Usage:      "与令牌关联的初始TTL。根据配置的最大TTL，令牌续订可能会超出此值。它被指定为带有后缀“30s”或“5m”的数字字符串。",
	})

	f.DurationVar(&DurationVar{
		Name:       "explicit-max-ttl",
		Target:     &c.flagExplicitMaxTTL,
		Completion: complete.PredictAnything,
		Usage:      "令牌的显式最大生存期。与正常的TTL不同，最大TTL是一个硬限制，不能超过。它被指定为带有后缀“30s”或“5m”的数字字符串。",
	})

	f.DurationVar(&DurationVar{
		Name:       "period",
		Target:     &c.flagPeriod,
		Completion: complete.PredictAnything,
		Usage:      "如果指定，每次续订都将使用给定的时间段。周期性令牌不会过期（除非还提供了-explicit-max-ttl）。设置此值需要sudo权限。它被指定为带有后缀“30s”或“5m”的数字字符串。",
	})

	f.BoolVar(&BoolVar{
		Name:    "renewable",
		Target:  &c.flagRenewable,
		Default: true,
		Usage:   "允许令牌更新到其最大TTL。",
	})

	f.BoolVar(&BoolVar{
		Name:    "orphan",
		Target:  &c.flagOrphan,
		Default: false,
		Usage:   "创建没有父级的令牌。这可以防止在创建令牌的令牌过期时吊销令牌。设置此值需要sudo权限。",
	})

	f.BoolVar(&BoolVar{
		Name:    "no-default-policy",
		Target:  &c.flagNoDefaultPolicy,
		Default: false,
		Usage:   "将“default”策略与此令牌的策略集分离。",
	})

	f.IntVar(&IntVar{
		Name:    "use-limit",
		Target:  &c.flagUseLimit,
		Default: 0,
		Usage:   "可使用此令牌的次数。最后一次使用后，令牌将自动吊销。默认情况下，令牌可以无限次使用，直到到期为止。",
	})

	f.StringVar(&StringVar{
		Name:    "role",
		Target:  &c.flagRole,
		Default: "",
		Usage:   "要创建令牌的角色的名称。指定-role可能会重写其他参数。本地验证的保管库令牌必须具有“auth/token/create/<role>”的权限。",
	})

	f.StringVar(&StringVar{
		Name:    "type",
		Target:  &c.flagType,
		Default: "service",
		Usage:   `要创建的令牌类型。可以是“service”或“batch”。`,
	})

	f.StringMapVar(&StringMapVar{
		Name:       "metadata",
		Target:     &c.flagMetadata,
		Completion: complete.PredictAnything,
		Usage:      "要与令牌关联的任意key=value元数据。使用令牌时，此元数据将显示在审核日志中。可以多次指定此选项以添加多个元数据。",
	})

	f.StringSliceVar(&StringSliceVar{
		Name:       "policy",
		Target:     &c.flagPolicies,
		Completion: c.PredictVaultPolicies(),
		Usage:      "要与此令牌关联的策略的名称。可以多次指定此选项以附加多个策略。",
	})

	return set
}

func (c *TokenCreateCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *TokenCreateCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *TokenCreateCommand) Run(args []string) int {
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

	if c.flagType == "batch" {
		c.flagRenewable = false
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	tcr := &api.TokenCreateRequest{
		ID:              c.flagID,
		Policies:        c.flagPolicies,
		Metadata:        c.flagMetadata,
		TTL:             c.flagTTL.String(),
		NoParent:        c.flagOrphan,
		NoDefaultPolicy: c.flagNoDefaultPolicy,
		DisplayName:     c.flagDisplayName,
		NumUses:         c.flagUseLimit,
		Renewable:       &c.flagRenewable,
		ExplicitMaxTTL:  c.flagExplicitMaxTTL.String(),
		Period:          c.flagPeriod.String(),
		Type:            c.flagType,
	}

	var secret *api.Secret
	if c.flagRole != "" {
		secret, err = client.Auth().Token().CreateWithRole(tcr, c.flagRole)
	} else {
		secret, err = client.Auth().Token().Create(tcr)
	}
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error creating token: %s", err))
		return 2
	}

	if c.flagField != "" {
		return PrintRawField(c.UI, secret, c.flagField)
	}

	return OutputSecret(c.UI, secret)
}
