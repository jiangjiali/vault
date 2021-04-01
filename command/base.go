package command

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/jiangjiali/vault/api"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jiangjiali/vault/command/token"
	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/errors"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/helper/namespace"
)

const (
	// maxLineLength is the maximum width of any line.
	maxLineLength int = 78

	// notSetValue是未设置值的标志值
	notSetValue = "(未设置)"
)

// reRemoveWhitespace is a regular expression for stripping whitespace from
// a string.
var reRemoveWhitespace = regexp.MustCompile(`[\s]+`)

type BaseCommand struct {
	UI cli.Ui

	flags     *FlagSets
	flagsOnce sync.Once

	flagAddress        string
	flagAgentAddress   string
	flagCACert         string
	flagCAPath         string
	flagClientCert     string
	flagClientKey      string
	flagNamespace      string
	flagNS             string
	flagPolicyOverride bool
	flagTLSServerName  string
	flagTLSSkipVerify  bool
	flagWrapTTL        time.Duration

	flagFormat           string
	flagField            string
	flagOutputCurlString bool

	flagMFA []string

	tokenHelper token.THelper

	client *api.Client
}

// Client returns the HTTP API client. The client is cached on the command to
// save performance on future calls.
func (b *BaseCommand) Client() (*api.Client, error) {
	// Read the test client if present
	if b.client != nil {
		return b.client, nil
	}

	config := api.DefaultConfig()

	if err := config.ReadEnvironment(); err != nil {
		return nil, errors.Wrap(err, "failed to read environment")
	}

	if b.flagAddress != "" {
		config.Address = b.flagAddress
	}
	if b.flagAgentAddress != "" {
		config.Address = b.flagAgentAddress
	}

	if b.flagOutputCurlString {
		config.OutputCurlString = b.flagOutputCurlString
	}

	// If we need custom TLS configuration, then set it
	if b.flagCACert != "" || b.flagCAPath != "" || b.flagClientCert != "" ||
		b.flagClientKey != "" || b.flagTLSServerName != "" || b.flagTLSSkipVerify {
		t := &api.TLSConfig{
			CACert:        b.flagCACert,
			CAPath:        b.flagCAPath,
			ClientCert:    b.flagClientCert,
			ClientKey:     b.flagClientKey,
			TLSServerName: b.flagTLSServerName,
			Insecure:      b.flagTLSSkipVerify,
		}
		config.ConfigureTLS(t)
	}

	// Build the client
	client, err := api.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	// Turn off retries on the CLI
	if os.Getenv(api.EnvVaultMaxRetries) == "" {
		client.SetMaxRetries(0)
	}

	// Set the wrapping function
	client.SetWrappingLookupFunc(b.DefaultWrappingLookupFunc)

	// Get the token if it came in from the environment
	xxtoken := client.Token()

	// If we don't have a token, check the token helper
	if xxtoken == "" {
		helper, err := b.TokenHelper()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get token helper")
		}
		xxtoken, err = helper.Get()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get token from token helper")
		}
	}

	// Set the token
	if xxtoken != "" {
		client.SetToken(xxtoken)
	}

	client.SetMFACreds(b.flagMFA)

	// flagNS takes precedence over flagNamespace. After resolution, point both
	// flags to the same value to be able to use them interchangeably anywhere.
	if b.flagNS != notSetValue {
		b.flagNamespace = b.flagNS
	}
	if b.flagNamespace != notSetValue {
		client.SetNamespace(namespace.Canonicalize(b.flagNamespace))
	}
	if b.flagPolicyOverride {
		client.SetPolicyOverride(b.flagPolicyOverride)
	}

	b.client = client

	return client, nil
}

// SetAddress sets the token helper on the command; useful for the demo server and other outside cases.
func (b *BaseCommand) SetAddress(addr string) {
	b.flagAddress = addr
}

// SetTokenHelper sets the token helper on the command.
func (b *BaseCommand) SetTokenHelper(th token.THelper) {
	b.tokenHelper = th
}

// TokenHelper returns the token helper attached to the command.
func (b *BaseCommand) TokenHelper() (token.THelper, error) {
	if b.tokenHelper != nil {
		return b.tokenHelper, nil
	}

	helper, err := DefaultTokenHelper()
	if err != nil {
		return nil, err
	}
	return helper, nil
}

// DefaultWrappingLookupFunc is the default wrapping function based on the
// CLI flag.
func (b *BaseCommand) DefaultWrappingLookupFunc(operation, path string) string {
	if b.flagWrapTTL != 0 {
		return b.flagWrapTTL.String()
	}

	return api.DefaultWrappingLookupFunc(operation, path)
}

type FlagSetBit uint

const (
	FlagSetNone FlagSetBit = 1 << iota
	FlagSetHTTP
	FlagSetOutputField
	FlagSetOutputFormat
)

// flagSet creates the flags for this command. The result is cached on the
// command to save performance on future calls.
func (b *BaseCommand) flagSet(bit FlagSetBit) *FlagSets {
	b.flagsOnce.Do(func() {
		set := NewFlagSets()

		// These flag sets will apply to all leaf subcommands.
		// TODO: Optional, but FlagSetHTTP can be safely removed from the individual
		// Flags() subcommands.
		bit = bit | FlagSetHTTP

		if bit&FlagSetHTTP != 0 {
			f := set.NewFlagSet("HTTP选项")

			addrStringVar := &StringVar{
				Name:       flagNameAddress,
				Target:     &b.flagAddress,
				EnvVar:     api.EnvVaultAddress,
				Completion: complete.PredictAnything,
				Usage:      "安全库服务器的地址。",
			}
			if b.flagAddress != "" {
				addrStringVar.Default = b.flagAddress
			} else {
				addrStringVar.Default = "https://127.0.0.1:8200"
			}
			f.StringVar(addrStringVar)

			agentAddrStringVar := &StringVar{
				Name:       "agent-address",
				Target:     &b.flagAgentAddress,
				EnvVar:     api.EnvVaultAgentAddr,
				Completion: complete.PredictAnything,
				Usage:      "代理的地址。",
			}
			f.StringVar(agentAddrStringVar)

			f.StringVar(&StringVar{
				Name:       flagNameCACert,
				Target:     &b.flagCACert,
				Default:    "",
				EnvVar:     api.EnvVaultCACert,
				Completion: complete.PredictFiles("*"),
				Usage:      "本地磁盘上指向单个PEM编码的CA证书的路径，以验证Vault服务器的SSL证书。这优先于 -ca-path。",
			})

			f.StringVar(&StringVar{
				Name:       flagNameCAPath,
				Target:     &b.flagCAPath,
				Default:    "",
				EnvVar:     api.EnvVaultCAPath,
				Completion: complete.PredictDirs("*"),
				Usage:      "本地磁盘上指向PEM编码的CA证书目录的路径，以验证Vault服务器的SSL证书。",
			})

			f.StringVar(&StringVar{
				Name:       flagNameClientCert,
				Target:     &b.flagClientCert,
				Default:    "",
				EnvVar:     api.EnvVaultClientCert,
				Completion: complete.PredictFiles("*"),
				Usage:      "本地磁盘上指向单个PEM编码的CA证书的路径，该证书用于对Vault服务器进行TLS身份验证。如果指定了此标志，则还需要-client-key。",
			})

			f.StringVar(&StringVar{
				Name:       flagNameClientKey,
				Target:     &b.flagClientKey,
				Default:    "",
				EnvVar:     api.EnvVaultClientKey,
				Completion: complete.PredictFiles("*"),
				Usage:      "本地磁盘上与 -client-cert 中的客户端证书匹配的单个PEM编码私钥的路径。",
			})

			f.StringVar(&StringVar{
				Name:       "namespace",
				Target:     &b.flagNamespace,
				Default:    notSetValue, // this can never be a real value
				EnvVar:     api.EnvVaultNamespace,
				Completion: complete.PredictAnything,
				Usage:      "用于命令的命名空间。不需要设置此选项，但允许使用相对路径。-ns 可以用作快捷方式。",
			})

			f.StringVar(&StringVar{
				Name:       "ns",
				Target:     &b.flagNS,
				Default:    notSetValue, // this can never be a real value
				Completion: complete.PredictAnything,
				Hidden:     true,
				Usage:      "-namespace的别名。这优先于-namespace。",
			})

			f.StringVar(&StringVar{
				Name:       "tls-server-name",
				Target:     &b.flagTLSServerName,
				Default:    "",
				EnvVar:     api.EnvVaultTLSServerName,
				Completion: complete.PredictAnything,
				Usage:      "通过TLS连接到安全库服务器时用作SNI主机的名称。",
			})

			f.BoolVar(&BoolVar{
				Name:    flagNameTLSSkipVerify,
				Target:  &b.flagTLSSkipVerify,
				Default: false,
				EnvVar:  api.EnvVaultSkipVerify,
				Usage:   "禁用TLS证书的验证。强烈建议不要使用此选项，因为它会降低与安全库服务器之间数据传输的安全性。",
			})

			f.BoolVar(&BoolVar{
				Name:    "policy-override",
				Target:  &b.flagPolicyOverride,
				Default: false,
				Usage:   "重写指定了软强制执行级别的哨兵策略",
			})

			f.DurationVar(&DurationVar{
				Name:       "wrap-ttl",
				Target:     &b.flagWrapTTL,
				Default:    0,
				EnvVar:     api.EnvVaultWrapTTL,
				Completion: complete.PredictAnything,
				Usage:      "用请求的TTL将响应包装在一个小隔间令牌中。可通过“vault unwrap”命令获得响应。TTL被指定为带有后缀“30s”或“5m”的数字字符串。",
			})

			f.StringSliceVar(&StringSliceVar{
				Name:       "mfa",
				Target:     &b.flagMFA,
				Default:    nil,
				EnvVar:     api.EnvVaultMFA,
				Completion: complete.PredictAnything,
				Usage:      "提供MFA凭据作为X-Vault-MFA标头的一部分。",
			})

			f.BoolVar(&BoolVar{
				Name:    "output-curl-string",
				Target:  &b.flagOutputCurlString,
				Default: false,
				Usage:   "不是执行请求，而是打印一个等效的cURL命令字符串并退出。",
			})

		}

		if bit&(FlagSetOutputField|FlagSetOutputFormat) != 0 {
			f := set.NewFlagSet("输出选项")

			if bit&FlagSetOutputField != 0 {
				f.StringVar(&StringVar{
					Name:       "field",
					Target:     &b.flagField,
					Default:    "",
					Completion: complete.PredictAnything,
					Usage:      "只打印具有给定名称的字段。指定此选项将优先于其他格式指令。结果将不会有一个尾随新线，使其成为管道到其他流程的理想选择。",
				})
			}

			if bit&FlagSetOutputFormat != 0 {
				f.StringVar(&StringVar{
					Name:       "format",
					Target:     &b.flagFormat,
					Default:    "table",
					EnvVar:     EnvVaultFormat,
					Completion: complete.PredictSet("table", "json"),
					Usage:      "以给定格式打印输出。有效格式为“table”、“json”。",
				})
			}
		}

		b.flags = set
	})

	return b.flags
}

// FlagSets is a group of flag sets.
type FlagSets struct {
	flagSets    []*FlagSet
	mainSet     *flag.FlagSet
	hiddens     map[string]struct{}
	completions complete.Flags
}

// NewFlagSets creates a new flag sets.
func NewFlagSets() *FlagSets {
	mainSet := flag.NewFlagSet("", flag.ContinueOnError)

	// Errors and usage are controlled by the CLI.
	mainSet.Usage = func() {}
	mainSet.SetOutput(ioutil.Discard)

	return &FlagSets{
		flagSets:    make([]*FlagSet, 0, 6),
		mainSet:     mainSet,
		hiddens:     make(map[string]struct{}),
		completions: complete.Flags{},
	}
}

// NewFlagSet creates a new flag set from the given flag sets.
func (f *FlagSets) NewFlagSet(name string) *FlagSet {
	flagSet := NewFlagSet(name)
	flagSet.mainSet = f.mainSet
	flagSet.completions = f.completions
	f.flagSets = append(f.flagSets, flagSet)
	return flagSet
}

// Completions returns the completions for this flag set.
func (f *FlagSets) Completions() complete.Flags {
	return f.completions
}

// Parse parses the given flags, returning any errors.
func (f *FlagSets) Parse(args []string) error {
	return f.mainSet.Parse(args)
}

// Parsed reports whether the command-line flags have been parsed.
func (f *FlagSets) Parsed() bool {
	return f.mainSet.Parsed()
}

// Args returns the remaining args after parsing.
func (f *FlagSets) Args() []string {
	return f.mainSet.Args()
}

// Visit visits the flags in lexicographical order, calling fn for each. It
// visits only those flags that have been set.
func (f *FlagSets) Visit(fn func(*flag.Flag)) {
	f.mainSet.Visit(fn)
}

// Help builds custom help for this command, grouping by flag set.
func (f *FlagSets) Help() string {
	var out bytes.Buffer

	for _, set := range f.flagSets {
		printFlagTitle(&out, set.name+":")
		set.VisitAll(func(f *flag.Flag) {
			// Skip any hidden flags
			if v, ok := f.Value.(FlagVisibility); ok && v.Hidden() {
				return
			}
			printFlagDetail(&out, f)
		})
	}

	return strings.TrimRight(out.String(), "\n")
}

// FlagSet is a grouped wrapper around a real flag set and a grouped flag set.
type FlagSet struct {
	name        string
	flagSet     *flag.FlagSet
	mainSet     *flag.FlagSet
	completions complete.Flags
}

// NewFlagSet creates a new flag set.
func NewFlagSet(name string) *FlagSet {
	return &FlagSet{
		name:    name,
		flagSet: flag.NewFlagSet(name, flag.ContinueOnError),
	}
}

// Name returns the name of this flag set.
func (f *FlagSet) Name() string {
	return f.name
}

func (f *FlagSet) Visit(fn func(*flag.Flag)) {
	f.flagSet.Visit(fn)
}

func (f *FlagSet) VisitAll(fn func(*flag.Flag)) {
	f.flagSet.VisitAll(fn)
}

// printFlagTitle prints a consistently-formatted title to the given writer.
func printFlagTitle(w io.Writer, s string) {
	fmt.Fprintf(w, "%s\n\n", s)
}

// printFlagDetail prints a single flag to the given writer.
func printFlagDetail(w io.Writer, f *flag.Flag) {
	// Check if the flag is hidden - do not print any flag detail or help output
	// if it is hidden.
	if h, ok := f.Value.(FlagVisibility); ok && h.Hidden() {
		return
	}

	// Check for a detailed example
	example := ""
	if t, ok := f.Value.(FlagExample); ok {
		example = t.Example()
	}

	if example != "" {
		fmt.Fprintf(w, "  -%s=<%s>\n", f.Name, example)
	} else {
		fmt.Fprintf(w, "  -%s\n", f.Name)
	}

	usage := reRemoveWhitespace.ReplaceAllString(f.Usage, " ")
	indented := wrapAtLengthWithPadding(usage, 6)
	fmt.Fprintf(w, "%s\n\n", indented)
}
