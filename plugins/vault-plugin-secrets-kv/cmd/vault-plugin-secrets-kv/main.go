package main

import (
	"os"

	"github.com/jiangjiali/vault/sdk/helper/hclutil/hclog"
	"github.com/jiangjiali/vault/sdk/helper/pluginutil"
	"github.com/jiangjiali/vault/sdk/plugin"

	kv "github.com/jiangjiali/vault/plugins/vault-plugin-secrets-kv"
)

func main() {
	apiClientMeta := &pluginutil.APIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := pluginutil.VaultPluginTLSProvider(tlsConfig)

	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: kv.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		logger := hclog.New(&hclog.LoggerOptions{})

		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
