package main

import (
	"log"
	"os"

	jwtauth "github.com/jiangjiali/vault/plugins/vault-plugin-auth-jwt"
	"github.com/jiangjiali/vault/sdk/helper/pluginutil"
	"github.com/jiangjiali/vault/sdk/plugin"
)

func main() {
	apiClientMeta := &pluginutil.APIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:]) // Ignore command, strictly parse flags

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := pluginutil.VaultPluginTLSProvider(tlsConfig)

	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: jwtauth.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
