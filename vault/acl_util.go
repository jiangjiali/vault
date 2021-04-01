// +build !enterprise

package vault

import (
	"context"

	"github.com/jiangjiali/vault/sdk/helper/identity"
	"github.com/jiangjiali/vault/sdk/logical"
)

func (c *Core) performEntPolicyChecks(ctx context.Context, acl *ACL, te *logical.TokenEntry, req *logical.Request, inEntity *identity.Entity, opts *PolicyCheckOpts, ret *AuthResults) {
	ret.Allowed = true
}
