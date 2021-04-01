package command

import (
	"bytes"
	"fmt"
	"github.com/jiangjiali/vault/api"
	"io"
	"os"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/fatih/structs"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/helper/password"
	"github.com/jiangjiali/vault/sdk/helper/pgpkeys"
)

var _ cli.Command = (*OperatorRekeyCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorRekeyCommand)(nil)

type OperatorRekeyCommand struct {
	*BaseCommand

	flagCancel       bool
	flagInit         bool
	flagKeyShares    int
	flagKeyThreshold int
	flagNonce        string
	flagPGPKeys      []string
	flagStatus       bool
	flagTarget       string
	flagVerify       bool

	// Backup options
	flagBackup         bool
	flagBackupDelete   bool
	flagBackupRetrieve bool

	testStdin io.Reader // for tests
}

func (c *OperatorRekeyCommand) Synopsis() string {
	return "生成新的解封密钥"
}

func (c *OperatorRekeyCommand) Help() string {
	helpText := `
使用: vault operator rekey [选项] [KEY]

  生成一组新的解封密钥。
  这可以选择更改密钥共享的总数或这些密钥共享的所需阈值，以重建主密钥。
  此操作是零停机时间，但它要求解封保险库并提供足够数量的现有解封密钥。

  解封键可以作为命令的参数直接在命令行上提供。
  如果key被指定为“-”，则命令将从stdin读取。
  如果命令TTY可用，则将显示文本提示。

  初始化密钥：

      $ vault operator rekey \
          -init \
          -key-shares=15 \
          -key-threshold=9

  使用PGP重新密钥和加密生成的解封密钥：

      $ vault operator rekey \
          -init \
          -key-shares=3 \
          -key-threshold=2 \
          -pgp-keys="keybase:hashicorp,keybase:jefferai,keybase:sethvargo"

  在安全库的核心中存储加密的PGP密钥：

      $ vault operator rekey \
          -init \
          -pgp-keys="..." \
          -backup

  检索备份的解封密钥：

      $ vault operator rekey -backup-retrieve

  删除备份的解封密钥：

      $ vault operator rekey -backup-delete

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *OperatorRekeyCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:    "init",
		Target:  &c.flagInit,
		Default: false,
		Usage:   "初始化重新键入操作。只有在没有正在进行的重新键入操作时，才能执行此操作。使用-key-shares和-key-threshold标志自定义新的密钥共享数和密钥阈值。",
	})

	f.BoolVar(&BoolVar{
		Name:    "cancel",
		Target:  &c.flagCancel,
		Default: false,
		Usage:   "重置重新键入进度。这将丢弃所有提交的解封密钥或配置。",
	})

	f.BoolVar(&BoolVar{
		Name:    "status",
		Target:  &c.flagStatus,
		Default: false,
		Usage:   "打印当前尝试的状态，而不提供解封密钥。",
	})

	f.IntVar(&IntVar{
		Name:       "key-shares",
		Aliases:    []string{"n"},
		Target:     &c.flagKeyShares,
		Default:    5,
		Completion: complete.PredictAnything,
		Usage:      "要分割生成的主密钥的密钥共享数进入。这个要生成的“解封密钥”数。",
	})

	f.IntVar(&IntVar{
		Name:       "key-threshold",
		Aliases:    []string{"t"},
		Target:     &c.flagKeyThreshold,
		Default:    3,
		Completion: complete.PredictAnything,
		Usage:      "重建主密钥所需的密钥共享数。这必须小于或等于 -key-shares。",
	})

	f.StringVar(&StringVar{
		Name:       "nonce",
		Target:     &c.flagNonce,
		Default:    "",
		EnvVar:     "",
		Completion: complete.PredictAnything,
		Usage:      "初始化时提供的Nonce值。每个解封密钥必须提供相同的nonce值。",
	})

	f.StringVar(&StringVar{
		Name:       "target",
		Target:     &c.flagTarget,
		Default:    "barrier",
		EnvVar:     "",
		Completion: complete.PredictSet("barrier", "recovery"),
		Usage:      "重新键入的目标。”恢复”仅在启用HSM支持时适用。",
	})

	f.BoolVar(&BoolVar{
		Name:    "verify",
		Target:  &c.flagVerify,
		Default: false,
		Usage:   "指示操作（-status、-cancel或提供密钥共享）将影响当前重新密钥尝试的验证。",
	})

	f.VarFlag(&VarFlag{
		Name:       "pgp-keys",
		Value:      (*pgpkeys.PubKeyFilesFlag)(&c.flagPGPKeys),
		Completion: complete.PredictAnything,
		Usage:      "磁盘上包含公共GPG密钥的文件路径的逗号分隔列表，或使用格式“keybase:<username>”的逗号分隔的Keybase用户名列表。提供后，生成的解封密钥将按此列表中指定的顺序进行加密和base64编码。",
	})

	f = set.NewFlagSet("Backup Options")

	f.BoolVar(&BoolVar{
		Name:    "backup",
		Target:  &c.flagBackup,
		Default: false,
		Usage:   "在保险库的核心中存储当前PGP加密解封密钥的备份。加密值可以在失败时恢复，也可以在成功后丢弃。有关详细信息，请参见-backup-delete和-backup-retrieve选项。此选项仅适用于现有的解封密钥是PGP加密的。",
	})

	f.BoolVar(&BoolVar{
		Name:    "backup-delete",
		Target:  &c.flagBackupDelete,
		Default: false,
		Usage:   "删除所有存储的备份解封密钥。",
	})

	f.BoolVar(&BoolVar{
		Name:    "backup-retrieve",
		Target:  &c.flagBackupRetrieve,
		Default: false,
		Usage:   "检索备份的解封密钥。仅当提供了PGP密钥且未删除备份时，此选项才可用。",
	})

	return set
}

func (c *OperatorRekeyCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictAnything
}

func (c *OperatorRekeyCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorRekeyCommand) Run(args []string) int {
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

	// Create the client
	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	switch {
	case c.flagBackupDelete:
		return c.backupDelete(client)
	case c.flagBackupRetrieve:
		return c.backupRetrieve(client)
	case c.flagCancel:
		return c.cancel(client)
	case c.flagInit:
		return c.init(client)
	case c.flagStatus:
		return c.status(client)
	default:
		// If there are no other flags, prompt for an unseal key.
		key := ""
		if len(args) > 0 {
			key = strings.TrimSpace(args[0])
		}
		return c.provide(client, key)
	}
}

// init starts the rekey process.
func (c *OperatorRekeyCommand) init(client *api.Client) int {
	// Handle the different API requests
	var fn func(*api.RekeyInitRequest) (*api.RekeyStatusResponse, error)
	switch strings.ToLower(strings.TrimSpace(c.flagTarget)) {
	case "barrier":
		fn = client.Sys().RekeyInit
	case "recovery", "hsm":
		fn = client.Sys().RekeyRecoveryKeyInit
	default:
		c.UI.Error(fmt.Sprintf("Unknown target: %s", c.flagTarget))
		return 1
	}

	// Make the request
	status, err := fn(&api.RekeyInitRequest{
		SecretShares:        c.flagKeyShares,
		SecretThreshold:     c.flagKeyThreshold,
		PGPKeys:             c.flagPGPKeys,
		Backup:              c.flagBackup,
		RequireVerification: c.flagVerify,
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error initializing rekey: %s", err))
		return 2
	}

	// Print warnings about recovery, etc.
	if len(c.flagPGPKeys) == 0 {
		if Format(c.UI) == "table" {
			c.UI.Warn(wrapAtLength(
				"WARNING! If you lose the keys after they are returned, there is no " +
					"recovery. Consider canceling this operation and re-initializing " +
					"with the -pgp-keys flag to protect the returned unseal keys along " +
					"with -backup to allow recovery of the encrypted keys in case of " +
					"emergency. You can delete the stored keys later using the -delete " +
					"flag."))
			c.UI.Output("")
		}
	}
	if len(c.flagPGPKeys) > 0 && !c.flagBackup {
		if Format(c.UI) == "table" {
			c.UI.Warn(wrapAtLength(
				"WARNING! You are using PGP keys for encrypted the resulting unseal " +
					"keys, but you did not enable the option to backup the keys to " +
					"Vault's core. If you lose the encrypted keys after they are " +
					"returned, you will not be able to recover them. Consider canceling " +
					"this operation and re-running with -backup to allow recovery of the " +
					"encrypted unseal keys in case of emergency. You can delete the " +
					"stored keys later using the -delete flag."))
			c.UI.Output("")
		}
	}

	// Provide the current status
	return c.printStatus(status)
}

// cancel is used to abort the rekey process.
func (c *OperatorRekeyCommand) cancel(client *api.Client) int {
	// Handle the different API requests
	var fn func() error
	switch strings.ToLower(strings.TrimSpace(c.flagTarget)) {
	case "barrier":
		fn = client.Sys().RekeyCancel
		if c.flagVerify {
			fn = client.Sys().RekeyVerificationCancel
		}
	case "recovery", "hsm":
		fn = client.Sys().RekeyRecoveryKeyCancel
		if c.flagVerify {
			fn = client.Sys().RekeyRecoveryKeyVerificationCancel
		}

	default:
		c.UI.Error(fmt.Sprintf("Unknown target: %s", c.flagTarget))
		return 1
	}

	// Make the request
	if err := fn(); err != nil {
		c.UI.Error(fmt.Sprintf("Error canceling rekey: %s", err))
		return 2
	}

	c.UI.Output("Success! Canceled rekeying (if it was started)")
	return 0
}

// provide prompts the user for the seal key and posts it to the update root
// endpoint. If this is the last unseal, this function outputs it.
func (c *OperatorRekeyCommand) provide(client *api.Client, key string) int {
	var statusFn func() (interface{}, error)
	var updateFn func(string, string) (interface{}, error)

	switch strings.ToLower(strings.TrimSpace(c.flagTarget)) {
	case "barrier":
		statusFn = func() (interface{}, error) {
			return client.Sys().RekeyStatus()
		}
		updateFn = func(s1 string, s2 string) (interface{}, error) {
			return client.Sys().RekeyUpdate(s1, s2)
		}
		if c.flagVerify {
			statusFn = func() (interface{}, error) {
				return client.Sys().RekeyVerificationStatus()
			}
			updateFn = func(s1 string, s2 string) (interface{}, error) {
				return client.Sys().RekeyVerificationUpdate(s1, s2)
			}
		}
	case "recovery", "hsm":
		statusFn = func() (interface{}, error) {
			return client.Sys().RekeyRecoveryKeyStatus()
		}
		updateFn = func(s1 string, s2 string) (interface{}, error) {
			return client.Sys().RekeyRecoveryKeyUpdate(s1, s2)
		}
		if c.flagVerify {
			statusFn = func() (interface{}, error) {
				return client.Sys().RekeyRecoveryKeyVerificationStatus()
			}
			updateFn = func(s1 string, s2 string) (interface{}, error) {
				return client.Sys().RekeyRecoveryKeyVerificationUpdate(s1, s2)
			}
		}
	default:
		c.UI.Error(fmt.Sprintf("Unknown target: %s", c.flagTarget))
		return 1
	}

	status, err := statusFn()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error getting rekey status: %s", err))
		return 2
	}

	var started bool
	var nonce string

	switch status.(type) {
	case *api.RekeyStatusResponse:
		stat := status.(*api.RekeyStatusResponse)
		started = stat.Started
		nonce = stat.Nonce
	case *api.RekeyVerificationStatusResponse:
		stat := status.(*api.RekeyVerificationStatusResponse)
		started = stat.Started
		nonce = stat.Nonce
	default:
		c.UI.Error("Unknown status type")
		return 1
	}

	// Verify a root token generation is in progress. If there is not one in
	// progress, return an error instructing the user to start one.
	if !started {
		c.UI.Error(wrapAtLength(
			"No rekey is in progress. Start a rekey process by running " +
				"\"vault operator rekey -init\"."))
		return 1
	}

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
		w := getWriterFromUI(c.UI)
		fmt.Fprintf(w, "Rekey operation nonce: %s\n", nonce)
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

	// Trim any whitespace from they key, especially since we might have
	// prompted the user for it.
	key = strings.TrimSpace(key)

	// Verify we have a nonce value
	if nonce == "" {
		c.UI.Error("Missing nonce value: specify it via the -nonce flag")
		return 1
	}

	// Provide the key, this may potentially complete the update
	resp, err := updateFn(key, nonce)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error posting unseal key: %s", err))
		return 2
	}

	var xxcomplete bool
	var mightContainUnsealKeys bool

	switch resp.(type) {
	case *api.RekeyUpdateResponse:
		xxcomplete = resp.(*api.RekeyUpdateResponse).Complete
		mightContainUnsealKeys = true
	case *api.RekeyVerificationUpdateResponse:
		xxcomplete = resp.(*api.RekeyVerificationUpdateResponse).Complete
	default:
		c.UI.Error("Unknown update response type")
		return 1
	}

	if !xxcomplete {
		return c.status(client)
	}

	if mightContainUnsealKeys {
		return c.printUnsealKeys(client, status.(*api.RekeyStatusResponse),
			resp.(*api.RekeyUpdateResponse))
	}

	c.UI.Output(wrapAtLength("Rekey verification successful. The rekey operation is complete and the new keys are now active."))
	return 0
}

// status is used just to fetch and dump the status.
func (c *OperatorRekeyCommand) status(client *api.Client) int {
	// Handle the different API requests
	var fn func() (interface{}, error)
	switch strings.ToLower(strings.TrimSpace(c.flagTarget)) {
	case "barrier":
		fn = func() (interface{}, error) {
			return client.Sys().RekeyStatus()
		}
		if c.flagVerify {
			fn = func() (interface{}, error) {
				return client.Sys().RekeyVerificationStatus()
			}
		}
	case "recovery", "hsm":
		fn = func() (interface{}, error) {
			return client.Sys().RekeyRecoveryKeyStatus()
		}
		if c.flagVerify {
			fn = func() (interface{}, error) {
				return client.Sys().RekeyRecoveryKeyVerificationStatus()
			}
		}
	default:
		c.UI.Error(fmt.Sprintf("Unknown target: %s", c.flagTarget))
		return 1
	}

	// Make the request
	status, err := fn()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading rekey status: %s", err))
		return 2
	}

	return c.printStatus(status)
}

// backupRetrieve retrieves the stored backup keys.
func (c *OperatorRekeyCommand) backupRetrieve(client *api.Client) int {
	// Handle the different API requests
	var fn func() (*api.RekeyRetrieveResponse, error)
	switch strings.ToLower(strings.TrimSpace(c.flagTarget)) {
	case "barrier":
		fn = client.Sys().RekeyRetrieveBackup
	case "recovery", "hsm":
		fn = client.Sys().RekeyRetrieveRecoveryBackup
	default:
		c.UI.Error(fmt.Sprintf("Unknown target: %s", c.flagTarget))
		return 1
	}

	// Make the request
	storedKeys, err := fn()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error retrieving rekey stored keys: %s", err))
		return 2
	}

	secret := &api.Secret{
		Data: structs.New(storedKeys).Map(),
	}

	return OutputSecret(c.UI, secret)
}

// backupDelete deletes the stored backup keys.
func (c *OperatorRekeyCommand) backupDelete(client *api.Client) int {
	// Handle the different API requests
	var fn func() error
	switch strings.ToLower(strings.TrimSpace(c.flagTarget)) {
	case "barrier":
		fn = client.Sys().RekeyDeleteBackup
	case "recovery", "hsm":
		fn = client.Sys().RekeyDeleteRecoveryBackup
	default:
		c.UI.Error(fmt.Sprintf("Unknown target: %s", c.flagTarget))
		return 1
	}

	// Make the request
	if err := fn(); err != nil {
		c.UI.Error(fmt.Sprintf("Error deleting rekey stored keys: %s", err))
		return 2
	}

	c.UI.Output("Success! Delete stored keys (if they existed)")
	return 0
}

// printStatus dumps the status to output
func (c *OperatorRekeyCommand) printStatus(in interface{}) int {
	var out []string
	out = append(out, "Key | Value")

	switch in.(type) {
	case *api.RekeyStatusResponse:
		status := in.(*api.RekeyStatusResponse)
		out = append(out, fmt.Sprintf("Nonce | %s", status.Nonce))
		out = append(out, fmt.Sprintf("Started | %t", status.Started))
		if status.Started {
			if status.Progress == status.Required {
				out = append(out, fmt.Sprintf("Rekey Progress | %d/%d (verification in progress)", status.Progress, status.Required))
			} else {
				out = append(out, fmt.Sprintf("Rekey Progress | %d/%d", status.Progress, status.Required))
			}
			out = append(out, fmt.Sprintf("New Shares | %d", status.N))
			out = append(out, fmt.Sprintf("New Threshold | %d", status.T))
			out = append(out, fmt.Sprintf("Verification Required | %t", status.VerificationRequired))
			if status.VerificationNonce != "" {
				out = append(out, fmt.Sprintf("Verification Nonce | %s", status.VerificationNonce))
			}
		}
		if len(status.PGPFingerprints) > 0 {
			out = append(out, fmt.Sprintf("PGP Fingerprints | %s", status.PGPFingerprints))
			out = append(out, fmt.Sprintf("Backup | %t", status.Backup))
		}
	case *api.RekeyVerificationStatusResponse:
		status := in.(*api.RekeyVerificationStatusResponse)
		out = append(out, fmt.Sprintf("Started | %t", status.Started))
		out = append(out, fmt.Sprintf("New Shares | %d", status.N))
		out = append(out, fmt.Sprintf("New Threshold | %d", status.T))
		out = append(out, fmt.Sprintf("Verification Nonce | %s", status.Nonce))
		out = append(out, fmt.Sprintf("Verification Progress | %d/%d", status.Progress, status.T))
	default:
		c.UI.Error("Unknown status type")
		return 1
	}

	switch Format(c.UI) {
	case "table":
		c.UI.Output(tableOutput(out, nil))
		return 0
	default:
		return OutputData(c.UI, in)
	}
}

func (c *OperatorRekeyCommand) printUnsealKeys(client *api.Client, status *api.RekeyStatusResponse, resp *api.RekeyUpdateResponse) int {
	switch Format(c.UI) {
	case "table":
	default:
		return OutputData(c.UI, resp)
	}

	// Space between the key prompt, if any, and the output
	c.UI.Output("")

	// Provide the keys
	var haveB64 bool
	if resp.KeysB64 != nil && len(resp.KeysB64) == len(resp.Keys) {
		haveB64 = true
	}
	for i, key := range resp.Keys {
		if len(resp.PGPFingerprints) > 0 {
			if haveB64 {
				c.UI.Output(fmt.Sprintf("Key %d fingerprint: %s; value: %s", i+1, resp.PGPFingerprints[i], resp.KeysB64[i]))
			} else {
				c.UI.Output(fmt.Sprintf("Key %d fingerprint: %s; value: %s", i+1, resp.PGPFingerprints[i], key))
			}
		} else {
			if haveB64 {
				c.UI.Output(fmt.Sprintf("Key %d: %s", i+1, resp.KeysB64[i]))
			} else {
				c.UI.Output(fmt.Sprintf("Key %d: %s", i+1, key))
			}
		}
	}

	c.UI.Output("")
	c.UI.Output(fmt.Sprintf("Operation nonce: %s", resp.Nonce))

	if len(resp.PGPFingerprints) > 0 && resp.Backup {
		c.UI.Output("")
		c.UI.Output(wrapAtLength(fmt.Sprintf(
			"The encrypted unseal keys are backed up to \"core/unseal-keys-backup\"" +
				"in the storage backend. Remove these keys at any time using " +
				"\"vault operator rekey -backup-delete\". Vault does not automatically " +
				"remove these keys.",
		)))
	}

	switch status.VerificationRequired {
	case false:
		c.UI.Output("")
		c.UI.Output(wrapAtLength(fmt.Sprintf(
			"Vault rekeyed with %d key shares and a key threshold of %d. Please "+
				"securely distribute the key shares printed above. When Vault is "+
				"re-sealed, restarted, or stopped, you must supply at least %d of "+
				"these keys to unseal it before it can start servicing requests.",
			status.N,
			status.T,
			status.T)))
	default:
		c.UI.Output("")
		c.UI.Output(wrapAtLength(fmt.Sprintf(
			"Vault has created a new key, split into %d key shares and a key threshold "+
				"of %d. These will not be active until after verification is complete. "+
				"Please securely distribute the key shares printed above. When Vault "+
				"is re-sealed, restarted, or stopped, you must supply at least %d of "+
				"these keys to unseal it before it can start servicing requests.",
			status.N,
			status.T,
			status.T)))
		c.UI.Output("")
		c.UI.Warn(wrapAtLength(
			"Again, these key shares are _not_ valid until verification is performed. " +
				"Do not lose or discard your current key shares until after verification " +
				"is complete or you will be unable to unseal Vault. If you cancel the " +
				"rekey process or seal Vault before verification is complete the new " +
				"shares will be discarded and the current shares will remain valid.",
		))
		c.UI.Output("")
		c.UI.Warn(wrapAtLength(
			"The current verification status, including initial nonce, is shown below.",
		))
		c.UI.Output("")

		c.flagVerify = true
		return c.status(client)
	}

	return 0
}
