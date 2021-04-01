package vault

import (
	"context"

	"github.com/jiangjiali/vault/sdk/helper/namespace"
)

var (
	NamespaceByID = namespaceByID
)

func namespaceByID(ctx context.Context, nsID string, c *Core) (*namespace.Namespace, error) {
	if nsID == namespace.RootNamespaceID {
		return namespace.RootNamespace, nil
	}
	return nil, namespace.ErrNoNamespace
}
