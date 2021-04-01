package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"io"
	"os"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
)

// LoginHandler is the interface that any auth handlers must implement to enable
// auth via the CLI.
type LoginHandler interface {
	Auth(*api.Client, map[string]string) (*api.Secret, error)
	Help() string
}

type LoginCommand struct {
	*BaseCommand

	Handlers map[string]LoginHandler

	flagMethod    string
	flagPath      string
	flagNoStore   bool
	flagNoPrint   bool
	flagTokenOnly bool

	testStdin io.Reader // for tests
}

func (c *LoginCommand) Synopsis() string {
	return "本地身份验证"
}

func (c *LoginCommand) Help() string {
	helpText := `
使用: vault login [选项] [AUTH K=V...]

  使用提供的参数验证要安全库存储的用户或计算机。
  成功的身份验证将生成安全库令牌-概念上类似于网站上的会话令牌。
  默认情况下，此令牌缓存在本地计算机上以备将来请求。

  默认的身份验证方法是“token”。
  如果未通过CLI提供，则Vault将提示输入。
  如果参数是“-”，则从stdin读取值。

  -method 标志允许使用其他身份验证方法，如userpass或cert。
  对于这些方法，可能需要额外的“K=V”对。
  例如，要对 userpass 认证方法进行身份验证，请执行以下操作：

      $ vault login -method=userpass username=my-username

  有关给定身份验证方法可用的配置参数列表的详细信息，
  请使用“服务器身份验证帮助类型”。
  也可以使用“服务器验证列表”查看已启用验证方法的列表。

  如果在非标准路径上启用了身份验证方法，则-method标志仍引用规范类型，
  但 -path 标志引用启用的路径。如果在“userpass-ent”处
  启用了 userpass 认证方法，请按如下方式进行身份验证：

      $ vault login -method=userpass -path=userpass-ent

  如果使用响应包装（通过 -wrap-ttl）请求身份验证，
  则返回的令牌将自动展开，除非：

    - 使用 -token-only 标志，在这种情况下，此命令将输出包装标记。

    - 使用 -no-store 标志，在这种情况下，此命令将输出包装令牌的详细信息。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *LoginCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputField | FlagSetOutputFormat)

	f := set.NewFlagSet("命令选项")

	f.StringVar(&StringVar{
		Name:       "method",
		Target:     &c.flagMethod,
		Default:    "token",
		Completion: c.PredictVaultAvailableAuths(),
		Usage:      "要使用的身份验证类型，例如“userpass”。注意这对应于类型，而不是启用的路径。使用-path指定启用身份验证的路径。",
	})

	f.StringVar(&StringVar{
		Name:       "path",
		Target:     &c.flagPath,
		Default:    "",
		Completion: c.PredictVaultAuths(),
		Usage:      "启用身份验证方法的安全库中的远程路径。默认为方法类型（例如userpass->userpass/）。",
	})

	f.BoolVar(&BoolVar{
		Name:    "no-store",
		Target:  &c.flagNoStore,
		Default: false,
		Usage:   "在身份验证之后，不要将令牌持久化到令牌助手（通常是本地文件系统），以备将来使用请求令牌将只显示在命令输出中。",
	})

	f.BoolVar(&BoolVar{
		Name:    "no-print",
		Target:  &c.flagNoPrint,
		Default: false,
		Usage:   "不显示令牌。令牌仍将存储到配置的令牌助手中。",
	})

	f.BoolVar(&BoolVar{
		Name:    "token-only",
		Target:  &c.flagTokenOnly,
		Default: false,
		Usage:   "只输出令牌，不进行验证。此标志是“-field=token -no-store”的快捷方式。将这些标志设置为其他值不会有任何影响。",
	})

	return set
}

func (c *LoginCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *LoginCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *LoginCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()

	// Set the right flags if the user requested token-only - this overrides
	// any previously configured values, as documented.
	if c.flagTokenOnly {
		c.flagNoStore = true
		c.flagField = "token"
	}

	if c.flagNoStore && c.flagNoPrint {
		c.UI.Error(wrapAtLength(
			"-no-store 与 -no-print 不能一起使用。"))
		return 1
	}

	// Get the auth method
	authMethod := sanitizePath(c.flagMethod)
	if authMethod == "" {
		authMethod = "token"
	}

	// If no path is specified, we default the path to the method type
	// or use the plugin name if it's a plugin
	authPath := c.flagPath
	if authPath == "" {
		authPath = ensureTrailingSlash(authMethod)
	}

	// Get the handler function
	authHandler, ok := c.Handlers[authMethod]
	if !ok {
		c.UI.Error(wrapAtLength(fmt.Sprintf(
			"Unknown auth method: %s. Use \"vault auth list\" to see the "+
				"complete list of auth methods. Additionally, some "+
				"auth methods are only available via the HTTP API.",
			authMethod)))
		return 1
	}

	// Pull our fake stdin if needed
	stdin := (io.Reader)(os.Stdin)
	if c.testStdin != nil {
		stdin = c.testStdin
	}

	// If the user provided a token, pass it along to the auth provider.
	if authMethod == "token" && len(args) > 0 && !strings.Contains(args[0], "=") {
		args = append([]string{"token=" + args[0]}, args[1:]...)
	}

	config, err := parseArgsDataString(stdin, args)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error parsing configuration: %s", err))
		return 1
	}

	// If the user did not specify a mount path, use the provided mount path.
	if config["mount"] == "" && authPath != "" {
		config["mount"] = authPath
	}

	// Create the client
	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	// Authenticate delegation to the auth handler
	secret, err := authHandler.Auth(client, config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error authenticating: %s", err))
		return 2
	}

	// Unset any previous token wrapping functionality. If the original request
	// was for a wrapped token, we don't want future requests to be wrapped.
	client.SetWrappingLookupFunc(func(string, string) string { return "" })

	// Recursively extract the token, handling wrapping
	unwrap := !c.flagTokenOnly && !c.flagNoStore
	secret, isWrapped, err := c.extractToken(client, secret, unwrap)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error extracting token: %s", err))
		return 2
	}
	if secret == nil {
		c.UI.Error("Vault returned an empty secret")
		return 2
	}

	// Handle special cases if the token was wrapped
	if isWrapped {
		if c.flagTokenOnly {
			return PrintRawField(c.UI, secret, "wrapping_token")
		}
		if c.flagNoStore {
			return OutputSecret(c.UI, secret)
		}
	}

	// If we got this far, verify we have authentication data before continuing
	if secret.Auth == nil {
		c.UI.Error(wrapAtLength(
			"安全库返回了一个机密，但该机密没有附加身份验证信息。这不应该发生，很可能是一个错误。"))
		return 2
	}

	// Pull the token itself out, since we don't need the rest of the auth
	// information anymore/.
	token := secret.Auth.ClientToken

	if !c.flagNoStore {
		// Grab the token helper so we can store
		tokenHelper, err := c.TokenHelper()
		if err != nil {
			c.UI.Error(wrapAtLength(fmt.Sprintf(
				"Error initializing token helper. Please verify that the token "+
					"helper is available and properly configured for your system. The "+
					"error was: %s", err)))
			return 1
		}

		// Store the token in the local client
		if err := tokenHelper.Store(token); err != nil {
			c.UI.Error(fmt.Sprintf("Error storing token: %s", err))
			c.UI.Error(wrapAtLength(
				"身份验证成功，但令牌未持久化。下面为您的记录显示了生成的令牌。") + "\n")
			OutputSecret(c.UI, secret)
			return 2
		}

		// Warn if the VAULT_TOKEN environment variable is set, as that will take
		// precedence. We output as a warning, so piping should still work since it
		// will be on a different stream.
		if os.Getenv("VAULT_TOKEN") != "" {
			c.UI.Warn(wrapAtLength("警告！已设置 VAULT_TOKEN 环境变量！"+
				"它优先于此命令设置的值。要使用此命令设置的值，"+
				"请取消设置 VAULT_TOKEN 环境变量或将其设置为下面显示的标记。") + "\n")
		}
	} else if !c.flagTokenOnly {
		// If token-only the user knows it won't be stored, so don't warn
		c.UI.Warn(wrapAtLength(
			"令牌未存储在令牌帮助程序中。设置VAULT_TOKEN环境变量或将下面的令牌与每个请求一起传递到VAULT。") + "\n")
	}

	if c.flagNoPrint {
		return 0
	}

	// If the user requested a particular field, print that out now since we
	// are likely piping to another process.
	if c.flagField != "" {
		return PrintRawField(c.UI, secret, c.flagField)
	}

	// Print some yay! text, but only in table mode.
	if Format(c.UI) == "table" {
		c.UI.Output(wrapAtLength(
			"成功！您现在已通过身份验证。下面显示的令牌信息已存储在令牌助手中。您不需要再次运行"+
				"\"vault login\" 。将来的安全库请求将自动使用此令牌。") + "\n")
	}

	return OutputSecret(c.UI, secret)
}

// extractToken extracts the token from the given secret, automatically
// unwrapping responses and handling error conditions if unwrap is true. The
// result also returns whether it was a wrapped response that was not unwrapped.
func (c *LoginCommand) extractToken(client *api.Client, secret *api.Secret, unwrap bool) (*api.Secret, bool, error) {
	switch {
	case secret == nil:
		return nil, false, fmt.Errorf("empty response from auth helper")

	case secret.Auth != nil:
		return secret, false, nil

	case secret.WrapInfo != nil:
		if secret.WrapInfo.WrappedAccessor == "" {
			return nil, false, fmt.Errorf("wrapped response does not contain a token")
		}

		if !unwrap {
			return secret, true, nil
		}

		client.SetToken(secret.WrapInfo.Token)
		secret, err := client.Logical().Unwrap("")
		if err != nil {
			return nil, false, err
		}
		return c.extractToken(client, secret, unwrap)

	default:
		return nil, false, fmt.Errorf("no auth or wrapping info in response")
	}
}
