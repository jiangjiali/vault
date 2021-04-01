package command

import (
	"fmt"
	"github.com/jiangjiali/vault/api"
	"strings"

	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/helper/pgpkeys"
)

var _ cli.Command = (*OperatorInitCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorInitCommand)(nil)

type OperatorInitCommand struct {
	*BaseCommand

	flagStatus          bool
	flagKeyShares       int
	flagKeyThreshold    int
	flagPGPKeys         []string
	flagRootTokenPGPKey string

	// Auto Unseal
	flagRecoveryShares    int
	flagRecoveryThreshold int
	flagRecoveryPGPKeys   []string
	flagStoredShares      int
}

const (
	defKeyShares    = 5
	defKeyThreshold = 3
)

func (c *OperatorInitCommand) Synopsis() string {
	return "初始化服务器"
}

func (c *OperatorInitCommand) Help() string {
	helpText := `
使用: vault operator init [选项]

  初始化安全库服务器。初始化是安全库存储后端准备接收数据的过程。
  由于安全库服务器在HA模式下共享同一存储后端，
  因此只需初始化一个安全库即可初始化存储后端。

  在初始化过程中，安全库生成内存中的主密钥，
  并应用Shamir的秘密共享算法将该主密钥分解为配置数量的密钥共享，
  以便这些密钥共享的可配置子集必须聚集在一起以重新生成主密钥。
  在安全库文档中，这些密钥通常称为“解封密钥”。

  无法对已初始化的安全库群集运行此命令。

  使用默认选项启动初始化：

      $ vault operator init

  初始化，但用PGP密钥加密解封密钥：

      $ vault operator init \
          -key-shares=3 \
          -key-threshold=2 \
          -pgp-keys="keybase:hashicorp,keybase:jefferai,keybase:sethvargo"

  使用PGP密钥加密初始根令牌：

      $ vault operator init -root-token-pgp-key="keybase:hashicorp"

` + c.Flags().Help()
	return strings.TrimSpace(helpText)
}

func (c *OperatorInitCommand) Flags() *FlagSets {
	set := c.flagSet(FlagSetHTTP | FlagSetOutputFormat)

	// Common Options
	f := set.NewFlagSet("命令选项")

	f.BoolVar(&BoolVar{
		Name:    "status",
		Target:  &c.flagStatus,
		Default: false,
		Usage:   "打印当前初始化状态。退出代码0表示安全库已初始化。退出代码1表示发生错误。退出代码为2表示平均值未初始化。",
	})

	f.IntVar(&IntVar{
		Name:       "key-shares",
		Aliases:    []string{"n"},
		Target:     &c.flagKeyShares,
		Default:    defKeyShares,
		Completion: complete.PredictAnything,
		Usage:      "要将生成的主密钥拆分为的密钥共享数。这是要生成的“解封密钥”的数量。",
	})

	f.IntVar(&IntVar{
		Name:       "key-threshold",
		Aliases:    []string{"t"},
		Target:     &c.flagKeyThreshold,
		Default:    defKeyThreshold,
		Completion: complete.PredictAnything,
		Usage:      "重建主密钥所需的密钥共享数。这必须小于或等于-密钥共享。",
	})

	f.VarFlag(&VarFlag{
		Name:       "pgp-keys",
		Value:      (*pgpkeys.PubKeyFilesFlag)(&c.flagPGPKeys),
		Completion: complete.PredictAnything,
		Usage:      "磁盘上包含公共GPG密钥的文件路径的逗号分隔列表，或使用格式“keybase:<username>”的逗号分隔的Keybase用户名列表。提供后，生成的解封密钥将按此列表中指定的顺序进行加密和base64编码。条目数必须与-密钥共享匹配，除非使用了-store共享。",
	})

	f.VarFlag(&VarFlag{
		Name:       "root-token-pgp-key",
		Value:      (*pgpkeys.PubKeyFileFlag)(&c.flagRootTokenPGPKey),
		Completion: complete.PredictAnything,
		Usage:      "磁盘上包含二进制或base64编码的公共GPG密钥的文件的路径。也可以使用“keybase:<username>”格式将其指定为Keybase用户名。当提供时，生成的根令牌将使用给定的公钥进行加密和base64编码。",
	})

	f.IntVar(&IntVar{
		Name:    "stored-shares",
		Target:  &c.flagStoredShares,
		Default: -1,
		Usage:   "已弃用：此标志不起任何作用。它将删除。",
	})

	// Auto Unseal Options
	f = set.NewFlagSet("Auto Unseal Options")

	f.IntVar(&IntVar{
		Name:       "recovery-shares",
		Target:     &c.flagRecoveryShares,
		Default:    5,
		Completion: complete.PredictAnything,
		Usage:      "要分割恢复密钥的密钥共享数进入。这个仅在自动解封模式下使用。",
	})

	f.IntVar(&IntVar{
		Name:       "recovery-threshold",
		Target:     &c.flagRecoveryThreshold,
		Default:    3,
		Completion: complete.PredictAnything,
		Usage:      "重建恢复密钥所需的密钥共享数。仅在自动解封模式下使用。",
	})

	f.VarFlag(&VarFlag{
		Name:       "recovery-pgp-keys",
		Value:      (*pgpkeys.PubKeyFilesFlag)(&c.flagRecoveryPGPKeys),
		Completion: complete.PredictAnything,
		Usage:      "其行为类似于 -pgp-keys，但对于恢复密钥共享。仅在自动解封模式下使用。",
	})

	return set
}

func (c *OperatorInitCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *OperatorInitCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorInitCommand) Run(args []string) int {
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

	if c.flagStoredShares != -1 {
		c.UI.Warn("-stored-shares has no effect and will be removed in Vault 1.3.\n")
	}

	// Build the initial init request
	initReq := &api.InitRequest{
		SecretShares:    c.flagKeyShares,
		SecretThreshold: c.flagKeyThreshold,
		PGPKeys:         c.flagPGPKeys,
		RootTokenPGPKey: c.flagRootTokenPGPKey,

		RecoveryShares:    c.flagRecoveryShares,
		RecoveryThreshold: c.flagRecoveryThreshold,
		RecoveryPGPKeys:   c.flagRecoveryPGPKeys,
	}

	client, err := c.Client()
	if err != nil {
		c.UI.Error(err.Error())
		return 2
	}

	// Check auto mode
	switch {
	case c.flagStatus:
		return c.status(client)
	default:
		return c.init(client, initReq)
	}
}

func (c *OperatorInitCommand) init(client *api.Client, req *api.InitRequest) int {
	resp, err := client.Sys().Init(req)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error initializing: %s", err))
		return 2
	}

	switch Format(c.UI) {
	case "table":
	default:
		return OutputData(c.UI, newMachineInit(req, resp))
	}

	for i, key := range resp.Keys {
		if resp.KeysB64 != nil && len(resp.KeysB64) == len(resp.Keys) {
			c.UI.Output(fmt.Sprintf("Unseal Key %d: %s", i+1, resp.KeysB64[i]))
		} else {
			c.UI.Output(fmt.Sprintf("Unseal Key %d: %s", i+1, key))
		}
	}
	for i, key := range resp.RecoveryKeys {
		if resp.RecoveryKeysB64 != nil && len(resp.RecoveryKeysB64) == len(resp.RecoveryKeys) {
			c.UI.Output(fmt.Sprintf("Recovery Key %d: %s", i+1, resp.RecoveryKeysB64[i]))
		} else {
			c.UI.Output(fmt.Sprintf("Recovery Key %d: %s", i+1, key))
		}
	}

	c.UI.Output("")
	c.UI.Output(fmt.Sprintf("Initial Root Token: %s", resp.RootToken))

	if len(resp.Keys) > 0 {
		c.UI.Output("")
		c.UI.Output(wrapAtLength(fmt.Sprintf(
			"Vault initialized with %d key shares and a key threshold of %d. Please "+
				"securely distribute the key shares printed above. When the Vault is "+
				"re-sealed, restarted, or stopped, you must supply at least %d of "+
				"these keys to unseal it before it can start servicing requests.",
			req.SecretShares,
			req.SecretThreshold,
			req.SecretThreshold)))

		c.UI.Output("")
		c.UI.Output(wrapAtLength(fmt.Sprintf(
			"Vault does not store the generated master key. Without at least %d "+
				"key to reconstruct the master key, Vault will remain permanently "+
				"sealed!",
			req.SecretThreshold)))

		c.UI.Output("")
		c.UI.Output(wrapAtLength(
			"It is possible to generate new unseal keys, provided you have a quorum " +
				"of existing unseal keys shares. See \"vault operator rekey\" for " +
				"more information."))
	} else {
		c.UI.Output("")
		c.UI.Output("Success! Vault is initialized")
	}

	if len(resp.RecoveryKeys) > 0 {
		c.UI.Output("")
		c.UI.Output(wrapAtLength(fmt.Sprintf(
			"Recovery key initialized with %d key shares and a key threshold of %d. "+
				"Please securely distribute the key shares printed above.",
			req.RecoveryShares,
			req.RecoveryThreshold)))
	}

	if len(resp.RecoveryKeys) > 0 && (req.SecretShares != defKeyShares || req.SecretThreshold != defKeyThreshold) {
		c.UI.Output("")
		c.UI.Warn(wrapAtLength(
			"WARNING! -key-shares and -key-threshold is ignored when " +
				"Auto Unseal is used. Use -recovery-shares and -recovery-threshold instead.",
		))
	}

	return 0
}

// status inspects the init status of vault and returns an appropriate error
// code and message.
func (c *OperatorInitCommand) status(client *api.Client) int {
	inited, err := client.Sys().InitStatus()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error checking init status: %s", err))
		return 1 // Normally we'd return 2, but 2 means something special here
	}

	if inited {
		c.UI.Output("Vault is initialized")
		return 0
	}

	c.UI.Output("Vault is not initialized")
	return 2
}

// machineInit is used to output information about the init command.
type machineInit struct {
	UnsealKeysB64     []string `json:"unseal_keys_b64"`
	UnsealKeysHex     []string `json:"unseal_keys_hex"`
	UnsealShares      int      `json:"unseal_shares"`
	UnsealThreshold   int      `json:"unseal_threshold"`
	RecoveryKeysB64   []string `json:"recovery_keys_b64"`
	RecoveryKeysHex   []string `json:"recovery_keys_hex"`
	RecoveryShares    int      `json:"recovery_keys_shares"`
	RecoveryThreshold int      `json:"recovery_keys_threshold"`
	RootToken         string   `json:"root_token"`
}

func newMachineInit(req *api.InitRequest, resp *api.InitResponse) *machineInit {
	init := &machineInit{}

	init.UnsealKeysHex = make([]string, len(resp.Keys))
	for i, v := range resp.Keys {
		init.UnsealKeysHex[i] = v
	}

	init.UnsealKeysB64 = make([]string, len(resp.KeysB64))
	for i, v := range resp.KeysB64 {
		init.UnsealKeysB64[i] = v
	}

	// If we don't get a set of keys back, it means that we are storing the keys,
	// so the key shares and threshold has been set to 1.
	if len(resp.Keys) == 0 {
		init.UnsealShares = 1
		init.UnsealThreshold = 1
	} else {
		init.UnsealShares = req.SecretShares
		init.UnsealThreshold = req.SecretThreshold
	}

	init.RecoveryKeysHex = make([]string, len(resp.RecoveryKeys))
	for i, v := range resp.RecoveryKeys {
		init.RecoveryKeysHex[i] = v
	}

	init.RecoveryKeysB64 = make([]string, len(resp.RecoveryKeysB64))
	for i, v := range resp.RecoveryKeysB64 {
		init.RecoveryKeysB64[i] = v
	}

	init.RecoveryShares = req.RecoveryShares
	init.RecoveryThreshold = req.RecoveryThreshold

	init.RootToken = resp.RootToken

	return init
}
