package logical

import (
	"time"

	"github.com/jiangjiali/vault/sdk/helper/sockaddr"
)

type TokenType uint8

const (
	// TokenTypeDefault means "use the default, if any, that is currently set
	// on the mount". If not set, results in a Service token.
	TokenTypeDefault TokenType = iota

	// TokenTypeService is a "normal" Vault token for long-lived services
	TokenTypeService

	// TokenTypeBatch is a batch token
	TokenTypeBatch

	// TokenTypeDefaultService, configured on a mount, means that if
	// TokenTypeDefault is sent back by the mount, create Service tokens
	TokenTypeDefaultService

	// TokenTypeDefaultBatch, configured on a mount, means that if
	// TokenTypeDefault is sent back by the mount, create Batch tokens
	TokenTypeDefaultBatch
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeDefault:
		return "default"
	case TokenTypeService:
		return "service"
	case TokenTypeBatch:
		return "batch"
	case TokenTypeDefaultService:
		return "default-service"
	case TokenTypeDefaultBatch:
		return "default-batch"
	default:
		panic("unreachable")
	}
}

// TokenEntry is used to represent a given token
type TokenEntry struct {
	Type TokenType `json:"type" mapstructure:"type" structs:"type" sentinel:""`

	// ID of this entry, generally a random UUID
	ID string `json:"id" mapstructure:"id" structs:"id" sentinel:""`

	// Accessor for this token, a random UUID
	Accessor string `json:"accessor" mapstructure:"accessor" structs:"accessor" sentinel:""`

	// Parent token, used for revocation trees
	Parent string `json:"parent" mapstructure:"parent" structs:"parent" sentinel:""`

	// Which named policies should be used
	Policies []string `json:"policies" mapstructure:"policies" structs:"policies"`

	// Used for audit trails, this is something like "auth/user/login"
	Path string `json:"path" mapstructure:"path" structs:"path"`

	// Used for auditing. This could include things like "source", "user", "ip"
	Meta map[string]string `json:"meta" mapstructure:"meta" structs:"meta" sentinel:"meta"`

	// Used for operators to be able to associate with the source
	DisplayName string `json:"display_name" mapstructure:"display_name" structs:"display_name"`

	// Used to restrict the number of uses (zero is unlimited). This is to
	// support one-time-tokens (generalized). There are a few special values:
	// if it's -1 it has run through its use counts and is executing its final
	// use; if it's -2 it is tainted, which means revocation is currently
	// running on it; and if it's -3 it's also tainted but revocation
	// previously ran and failed, so this hints the tidy function to try it
	// again.
	NumUses int `json:"num_uses" mapstructure:"num_uses" structs:"num_uses"`

	// Time of token creation
	CreationTime int64 `json:"creation_time" mapstructure:"creation_time" structs:"creation_time" sentinel:""`

	// Duration set when token was created
	TTL time.Duration `json:"ttl" mapstructure:"ttl" structs:"ttl" sentinel:""`

	// Explicit maximum TTL on the token
	ExplicitMaxTTL time.Duration `json:"explicit_max_ttl" mapstructure:"explicit_max_ttl" structs:"explicit_max_ttl" sentinel:""`

	// If set, the role that was used for parameters at creation time
	Role string `json:"role" mapstructure:"role" structs:"role"`

	// If set, the period of the token. This is only used when created directly
	// through the create endpoint; periods managed by roles or other auth
	// backends are subject to those renewal rules.
	Period time.Duration `json:"period" mapstructure:"period" structs:"period" sentinel:""`

	// These are the deprecated fields
	DisplayNameDeprecated    string        `json:"DisplayName" mapstructure:"DisplayName" structs:"DisplayName" sentinel:""`
	NumUsesDeprecated        int           `json:"NumUses" mapstructure:"NumUses" structs:"NumUses" sentinel:""`
	CreationTimeDeprecated   int64         `json:"CreationTime" mapstructure:"CreationTime" structs:"CreationTime" sentinel:""`
	ExplicitMaxTTLDeprecated time.Duration `json:"ExplicitMaxTTL" mapstructure:"ExplicitMaxTTL" structs:"ExplicitMaxTTL" sentinel:""`

	EntityID string `json:"entity_id" mapstructure:"entity_id" structs:"entity_id"`

	// The set of CIDRs that this token can be used with
	BoundCIDRs []*sockaddr.SockAddrMarshaler `json:"bound_cidrs" sentinel:""`

	// NamespaceID is the identifier of the namespace to which this token is
	// confined to. Do not return this value over the API when the token is
	// being looked up.
	NamespaceID string `json:"namespace_id" mapstructure:"namespace_id" structs:"namespace_id" sentinel:""`

	// PublicID is the identifier of the public storage belonging to this
	// token
	PublicID string `json:"public_id" mapstructure:"public_id" structs:"public_id" sentinel:""`
}

func (te *TokenEntry) SentinelGet(key string) (interface{}, error) {
	if te == nil {
		return nil, nil
	}
	switch key {
	case "policies":
		return te.Policies, nil

	case "path":
		return te.Path, nil

	case "display_name":
		return te.DisplayName, nil

	case "num_uses":
		return te.NumUses, nil

	case "role":
		return te.Role, nil

	case "entity_id":
		return te.EntityID, nil

	case "period":
		return te.Period, nil

	case "period_seconds":
		return int64(te.Period.Seconds()), nil

	case "explicit_max_ttl":
		return te.ExplicitMaxTTL, nil

	case "explicit_max_ttl_seconds":
		return int64(te.ExplicitMaxTTL.Seconds()), nil

	case "creation_ttl":
		return te.TTL, nil

	case "creation_ttl_seconds":
		return int64(te.TTL.Seconds()), nil

	case "creation_time":
		return time.Unix(te.CreationTime, 0).Format(time.RFC3339Nano), nil

	case "creation_time_unix":
		return time.Unix(te.CreationTime, 0), nil

	case "meta", "metadata":
		return te.Meta, nil

	case "type":
		teType := te.Type
		switch teType {
		case TokenTypeBatch, TokenTypeService:
		case TokenTypeDefault:
			teType = TokenTypeService
		default:
			return "unknown", nil
		}
		return teType.String(), nil
	}

	return nil, nil
}

func (te *TokenEntry) SentinelKeys() []string {
	return []string{
		"period",
		"period_seconds",
		"explicit_max_ttl",
		"explicit_max_ttl_seconds",
		"creation_ttl",
		"creation_ttl_seconds",
		"creation_time",
		"creation_time_unix",
		"meta",
		"metadata",
		"type",
	}
}