package seal

import (
	"fmt"

	"github.com/jiangjiali/vault/command/server"
	log "github.com/jiangjiali/vault/sdk/helper/hclutil/hclog"
	"github.com/jiangjiali/vault/vault"
	"github.com/jiangjiali/vault/vault/seal"
)

var (
	ConfigureSeal = configureSeal
)

func configureSeal(configSeal *server.Seal, infoKeys *[]string, info *map[string]string, logger log.Logger, inseal vault.Seal) (outseal vault.Seal, err error) {
	switch configSeal.Type {
	case seal.Transit:
		return configureTransitSeal(configSeal, infoKeys, info, logger)

	case seal.PKCS11:
		return nil, fmt.Errorf("seal type 'pkcs11' requires the Vault Enterprise HSM binary")

	case seal.Shamir:
		return inseal, nil

	default:
		return nil, fmt.Errorf("unknown seal type %q", configSeal.Type)
	}
}
