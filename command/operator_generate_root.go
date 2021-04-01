package command

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/jiangjiali/vault/api"
	"io"
	"os"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/base62"
	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/errwrap"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/helper/password"
	"github.com/jiangjiali/vault/sdk/helper/pgpkeys"
	"github.com/jiangjiali/vault/sdk/helper/xor"
	"github.com/jiangjiali/vault/sdk/helper/xxuuid"
)

var _ cli.Command = (*OperatorGenerateRootCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorGenerateRootCommand)(nil)

type OperatorGenerateRootCommand struct {
	*BaseCommand

	flagInit        bool
	flagCancel      bool
	flagStatus      bool
	flagDecode      string
	flagOTP         string
	flagPGPKey      string
	flagNonce       string
	flagGenerateOTP bool
	flagDRToken     bool

	testStdin io.Reader // for tests
}

func (c *OperatorGenerateRootCommand) Synopsis() string {
	return "生成新的根令牌"
}

func (c *OperatorGenerateRootCommand) Help() string {
	helpText := `
使用: vault operator generate-root [选项] [KEY]

  通过合并法定股东人数生成新的根令牌。
  必须提供以下选项之一才能启动根令牌生成：

    - 通过“-otp”标志提供的base64编码的一次性密码（OTP）。
      使用“-generate-otp”标志生成可用值。
      结果标记在返回时与此值进行异或。
      使用“-decode”标志输出最终值。

    - 在“-pgp-key”标志中包含PGP密钥或基础密钥用户名的文件。
      生成的令牌使用此公钥加密。

  解封键可以作为命令的参数直接在命令行上提供。
  如果key被指定为“-”，则命令将从stdin读取。
  如果命令TTY可用，则将显示文本提示。

  为最终令牌生成OTP代码：

      $ vault operator generate-root -generate-otp

  启动根令牌生成：

      $ vault operator generate-root -init -otp="..."
      $ vault operator generate-root -init -pgp-key="..."

  输入解封密钥以生成根令牌：

      $ vault operator generate-root -otp="..."

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *OperatorGenerateRootCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:       "init",
		Target:     &c.flagInit,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "开始生成根令牌。这只能在当前没有正在进行的情况下进行。",
	})

	f.BoolVar(&BoolVar{
		Name:       "cancel",
		Target:     &c.flagCancel,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "重置根令牌生成进度。这将丢弃所有提交的解封密钥或配置。",
	})

	f.BoolVar(&BoolVar{
		Name:       "status",
		Target:     &c.flagStatus,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "打印当前尝试的状态，而不提供解封密钥。",
	})

	f.StringVar(&StringVar{
		Name:       "decode",
		Target:     &c.flagDecode,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "要解码的值；设置此值将触发解码操作。",
	})

	f.BoolVar(&BoolVar{
		Name:       "generate-otp",
		Target:     &c.flagGenerateOTP,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "生成并打印适合与“-init”标志一起使用的高熵一次性密码（OTP）。",
	})

	f.BoolVar(&BoolVar{
		Name:       "dr-token",
		Target:     &c.flagDRToken,
		Default:    false,
		EnvVar:     "",
		Completion: complete.PredictNothing,
		Usage:      "设置此标志以在DR操作令牌上生成根操作。",
	})

	f.StringVar(&StringVar{
		Name:       "otp",
		Target:     &c.flagOTP,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "与“-decode”或“-init”一起使用的OTP代码。",
	})

	f.VarFlag(&VarFlag{
		Name:       "pgp-key",
		Value:      (*pgpkeys.PubKeyFileFlag)(&c.flagPGPKey),
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "磁盘上包含二进制或base64编码的公共GPG密钥的文件的路径。也可以使用“keybase:<username>”格式将其指定为Keybase用户名。当提供时，生成的根令牌将使用给定的公钥进行加密和base64编码。",
	})

	f.StringVar(&StringVar{
		Name:       "nonce",
		Target:     &c.flagNonce,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "初始化时提供的Nonce值。每个解封密钥必须提供相同的nonce值。",
	})

	return set
}

func (c *OperatorGenerateRootCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *OperatorGenerateRootCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorGenerateRootCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = f.Args()
	if len(args) > 1 {
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 0-1, got %d)", len(args)))
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	switch {
	case c.flagGenerateOTP:
		otp, code := c.generateOTP(client, c.flagDRToken)
		if code == 0 {
			return PrintRaw(c.UI, otp)
		}
		return code
	case c.flagDecode != "":
		return c.decode(client, c.flagDecode, c.flagOTP, c.flagDRToken)
	case c.flagCancel:
		return c.cancel(client, c.flagDRToken)
	case c.flagInit:
		return c.init(client, c.flagOTP, c.flagPGPKey, c.flagDRToken)
	case c.flagStatus:
		return c.status(client, c.flagDRToken)
	default:
		// If there are no other flags, prompt for an unseal key.
		key := ""
		if len(args) > 0 {
			key = strings.TrimSpace(args[0])
		}
		return c.provide(client, key, c.flagDRToken)
	}
}

// generateOTP generates a suitable OTP code for generating a root token.
func (c *OperatorGenerateRootCommand) generateOTP(client *api.Client, drToken bool) (string, int) {
	f := client.Sys().GenerateRootStatus
	if drToken {
		f = client.Sys().GenerateDROperationTokenStatus
	}
	status, err := f()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error getting root generation status: %s", err))
		return "", 2
	}

	switch status.OTPLength {
	case 0:
		// This is the fallback case
		buf := make([]byte, 16)
		readLen, err := rand.Read(buf)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error reading random bytes: %s", err))
			return "", 2
		}

		if readLen != 16 {
			c.UI.Error(fmt.Sprintf("Read %d bytes when we should have read 16", readLen))
			return "", 2
		}

		return base64.StdEncoding.EncodeToString(buf), 0

	default:
		otp, err := base62.Random(status.OTPLength)
		if err != nil {
			c.UI.Error(errwrap.Wrapf("Error reading random bytes: {{err}}", err).Error())
			return "", 2
		}

		return otp, 0
	}
}

// decode decodes the given value using the otp.
func (c *OperatorGenerateRootCommand) decode(client *api.Client, encoded, otp string, drToken bool) int {
	if encoded == "" {
		c.UI.Error("Missing encoded value: use -decode=<string> to supply it")
		return 1
	}
	if otp == "" {
		c.UI.Error("Missing otp: use -otp to supply it")
		return 1
	}

	f := client.Sys().GenerateRootStatus
	if drToken {
		f = client.Sys().GenerateDROperationTokenStatus
	}
	status, err := f()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error getting root generation status: %s", err))
		return 2
	}

	switch status.OTPLength {
	case 0:
		// Backwards compat
		tokenBytes, err := xor.XORBase64(encoded, otp)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error xoring token: %s", err))
			return 1
		}

		token, err := xxuuid.FormatUUID(tokenBytes)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error formatting base64 token value: %s", err))
			return 1
		}

		return PrintRaw(c.UI, strings.TrimSpace(token))

	default:
		tokenBytes, err := base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			c.UI.Error(errwrap.Wrapf("Error decoding base64'd token: {{err}}", err).Error())
			return 1
		}

		tokenBytes, err = xor.XORBytes(tokenBytes, []byte(otp))
		if err != nil {
			c.UI.Error(errwrap.Wrapf("Error xoring token: {{err}}", err).Error())
			return 1
		}

		return PrintRaw(c.UI, string(tokenBytes))
	}
}

// init is used to start the generation process
func (c *OperatorGenerateRootCommand) init(client *api.Client, otp, pgpKey string, drToken bool) int {
	// Validate incoming fields. Either OTP OR PGP keys must be supplied.
	if otp != "" && pgpKey != "" {
		c.UI.Error("Error initializing: cannot specify both -otp and -pgp-key")
		return 1
	}

	// Start the root generation
	f := client.Sys().GenerateRootInit
	if drToken {
		f = client.Sys().GenerateDROperationTokenInit
	}
	status, err := f(otp, pgpKey)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error initializing root generation: %s", err))
		return 2
	}

	switch Format(c.UI) {
	case "table":
		return c.printStatus(status)
	default:
		return OutputData(c.UI, status)
	}
}

// provide prompts the user for the seal key and posts it to the update root
// endpoint. If this is the last unseal, this function outputs it.
func (c *OperatorGenerateRootCommand) provide(client *api.Client, key string, drToken bool) int {
	f := client.Sys().GenerateRootStatus
	if drToken {
		f = client.Sys().GenerateDROperationTokenStatus
	}
	status, err := f()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error getting root generation status: %s", err))
		return 2
	}

	// Verify a root token generation is in progress. If there is not one in
	// progress, return an error instructing the user to start one.
	if !status.Started {
		c.UI.Error(wrapAtLength(
			"No root generation is in progress. Start a root generation by " +
				"running \"vault operator generate-root -init\"."))
		c.UI.Warn(wrapAtLength(fmt.Sprintf(
			"If starting root generation using the OTP method and generating "+
				"your own OTP, the length of the OTP string needs to be %d "+
				"characters in length.", status.OTPLength)))
		return 1
	}

	var nonce string

	switch key {
	case "-": // Read from stdin
		nonce = c.flagNonce

		// Pull our fake stdin if needed
		stdin := (io.Reader)(os.Stdin)
		if c.testStdin != nil {
			stdin = c.testStdin
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, stdin); err != nil {
			c.UI.Error(fmt.Sprintf("Failed to read from stdin: %s", err))
			return 1
		}

		key = buf.String()
	case "": // Prompt using the tty
		// Nonce value is not required if we are prompting via the terminal
		nonce = status.Nonce

		w := getWriterFromUI(c.UI)
		fmt.Fprintf(w, "Operation nonce: %s\n", nonce)
		fmt.Fprintf(w, "Unseal Key (will be hidden): ")
		key, err = password.Read(os.Stdin)
		fmt.Fprintf(w, "\n")
		if err != nil {
			if err == password.ErrInterrupted {
				c.UI.Error("user canceled")
				return 1
			}

			c.UI.Error(wrapAtLength(fmt.Sprintf("An error occurred attempting to "+
				"ask for the unseal key. The raw error message is shown below, but "+
				"usually this is because you attempted to pipe a value into the "+
				"command or you are executing outside of a terminal (tty). If you "+
				"want to pipe the value, pass \"-\" as the argument to read from "+
				"stdin. The raw error was: %s", err)))
			return 1
		}
	default: // Supplied directly as an arg
		nonce = c.flagNonce
	}

	// Trim any whitespace from they key, especially since we might have prompted
	// the user for it.
	key = strings.TrimSpace(key)

	// Verify we have a nonce value
	if nonce == "" {
		c.UI.Error("Missing nonce value: specify it via the -nonce flag")
		return 1
	}

	// Provide the key, this may potentially complete the update
	fUpd := client.Sys().GenerateRootUpdate
	if drToken {
		fUpd = client.Sys().GenerateDROperationTokenUpdate
	}
	status, err = fUpd(key, nonce)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error posting unseal key: %s", err))
		return 2
	}
	switch Format(c.UI) {
	case "table":
		return c.printStatus(status)
	default:
		return OutputData(c.UI, status)
	}
}

// cancel cancels the root token generation
func (c *OperatorGenerateRootCommand) cancel(client *api.Client, drToken bool) int {
	f := client.Sys().GenerateRootCancel
	if drToken {
		f = client.Sys().GenerateDROperationTokenCancel
	}
	if err := f(); err != nil {
		c.UI.Error(fmt.Sprintf("Error canceling root token generation: %s", err))
		return 2
	}
	c.UI.Output("Success! Root token generation canceled (if it was started)")
	return 0
}

// status is used just to fetch and dump the status
func (c *OperatorGenerateRootCommand) status(client *api.Client, drToken bool) int {
	f := client.Sys().GenerateRootStatus
	if drToken {
		f = client.Sys().GenerateDROperationTokenStatus
	}
	status, err := f()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error getting root generation status: %s", err))
		return 2
	}
	switch Format(c.UI) {
	case "table":
		return c.printStatus(status)
	default:
		return OutputData(c.UI, status)
	}
}

// printStatus dumps the status to output
func (c *OperatorGenerateRootCommand) printStatus(status *api.GenerateRootStatusResponse) int {
	var out []string
	out = append(out, fmt.Sprintf("Nonce | %s", status.Nonce))
	out = append(out, fmt.Sprintf("Started | %t", status.Started))
	out = append(out, fmt.Sprintf("Progress | %d/%d", status.Progress, status.Required))
	out = append(out, fmt.Sprintf("Complete | %t", status.Complete))
	if status.PGPFingerprint != "" {
		out = append(out, fmt.Sprintf("PGP Fingerprint | %s", status.PGPFingerprint))
	}
	switch {
	case status.EncodedToken != "":
		out = append(out, fmt.Sprintf("Encoded Token | %s", status.EncodedToken))
	case status.EncodedRootToken != "":
		out = append(out, fmt.Sprintf("Encoded Root Token | %s", status.EncodedRootToken))
	}
	if status.OTP != "" {
		c.UI.Warn(wrapAtLength("A One-Time-Password has been generated for you and is shown in the OTP field. You will need this value to decode the resulting root token, so keep it safe."))
		out = append(out, fmt.Sprintf("OTP | %s", status.OTP))
	}
	if status.OTPLength != 0 {
		out = append(out, fmt.Sprintf("OTP Length | %d", status.OTPLength))
	}

	output := columnOutput(out, nil)
	c.UI.Output(output)
	return 0
}
