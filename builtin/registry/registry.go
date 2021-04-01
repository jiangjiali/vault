package registry

import (
	"github.com/jiangjiali/vault/sdk/helper/consts"
	"github.com/jiangjiali/vault/sdk/logical"

	credUserpass "github.com/jiangjiali/vault/builtin/credential/userpass"
	credJWT "github.com/jiangjiali/vault/plugins/vault-plugin-auth-jwt"

	logicalTotp "github.com/jiangjiali/vault/builtin/logical/totp"
	logicalTransit "github.com/jiangjiali/vault/builtin/logical/transit"
	logicalKv "github.com/jiangjiali/vault/plugins/vault-plugin-secrets-kv"
)

// Registry is inherently thread-safe because it's immutable.
// Thus, rather than creating multiple instances of it, we only need one.
var Registry = newRegistry()

// BuiltinFactory is the func signature that should be returned by
// the plugin's New() func.
type BuiltinFactory func() (interface{}, error)

func newRegistry() *registry {
	return &registry{
		credentialBackends: map[string]logical.Factory{
			"jwt":      credJWT.Factory,
			"oidc":     credJWT.Factory,
			"userpass": credUserpass.Factory,
		},
		logicalBackends: map[string]logical.Factory{
			"kv":      logicalKv.Factory,
			"totp":    logicalTotp.Factory,
			"transit": logicalTransit.Factory,
		},
	}
}

type registry struct {
	credentialBackends map[string]logical.Factory
	logicalBackends    map[string]logical.Factory
}

// Get returns the BuiltinFactory func for a particular backend plugin
// from the plugins map.
func (r *registry) Get(name string, pluginType consts.PluginType) (func() (interface{}, error), bool) {
	switch pluginType {
	case consts.PluginTypeCredential:
		f, ok := r.credentialBackends[name]
		return toFunc(f), ok
	case consts.PluginTypeSecrets:
		f, ok := r.logicalBackends[name]
		return toFunc(f), ok
	default:
		return nil, false
	}
}

// Keys returns the list of plugin names that are considered builtin plugins.
func (r *registry) Keys(pluginType consts.PluginType) []string {
	var keys []string
	switch pluginType {
	case consts.PluginTypeCredential:
		for key := range r.credentialBackends {
			keys = append(keys, key)
		}
	case consts.PluginTypeSecrets:
		for key := range r.logicalBackends {
			keys = append(keys, key)
		}
	}
	return keys
}

func (r *registry) Contains(name string, pluginType consts.PluginType) bool {
	for _, key := range r.Keys(pluginType) {
		if key == name {
			return true
		}
	}
	return false
}

func toFunc(ifc interface{}) func() (interface{}, error) {
	return func() (interface{}, error) {
		return ifc, nil
	}
}
