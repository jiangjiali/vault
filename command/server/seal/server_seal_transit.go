package seal

import (
	"github.com/jiangjiali/vault/command/server"
	"github.com/jiangjiali/vault/sdk/helper/errwrap"
	log "github.com/jiangjiali/vault/sdk/helper/hclutil/hclog"
	"github.com/jiangjiali/vault/sdk/logical"
	"github.com/jiangjiali/vault/vault"
	"github.com/jiangjiali/vault/vault/seal/transit"
)

func configureTransitSeal(configSeal *server.Seal, infoKeys *[]string, info *map[string]string, logger log.Logger) (vault.Seal, error) {
	transitSeal := transit.NewSeal(logger)
	sealInfo, err := transitSeal.SetConfig(configSeal.Config)
	if err != nil {
		// If the error is any other than logical.KeyNotFoundError, return the error
		if !errwrap.ContainsType(err, new(logical.KeyNotFoundError)) {
			return nil, err
		}
	}
	autoseal := vault.NewAutoSeal(transitSeal)
	if sealInfo != nil {
		*infoKeys = append(*infoKeys, "Seal Type", "Transit Address", "Transit Mount Path", "Transit Key Name")
		(*info)["Seal Type"] = configSeal.Type
		(*info)["Transit Address"] = sealInfo["address"]
		(*info)["Transit Mount Path"] = sealInfo["mount_path"]
		(*info)["Transit Key Name"] = sealInfo["key_name"]
		if namespace, ok := sealInfo["namespace"]; ok {
			*infoKeys = append(*infoKeys, "Transit Namespace")
			(*info)["Transit Namespace"] = namespace
		}
	}
	return autoseal, nil
}
