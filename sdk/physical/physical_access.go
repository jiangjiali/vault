package physical

import (
	"context"
)

// PhysicalAccess is a wrapper around physical.Backend that allows Core to
// expose its physical storage operations through PhysicalAccess() while
// restricting the ability to modify Core.physical itself.
type PAccess struct {
	physical Backend
}

var _ Backend = (*PAccess)(nil)

func NewPhysicalAccess(physical Backend) *PAccess {
	return &PAccess{physical: physical}
}

func (p *PAccess) Put(ctx context.Context, entry *Entry) error {
	return p.physical.Put(ctx, entry)
}

func (p *PAccess) Get(ctx context.Context, key string) (*Entry, error) {
	return p.physical.Get(ctx, key)
}

func (p *PAccess) Delete(ctx context.Context, key string) error {
	return p.physical.Delete(ctx, key)
}

func (p *PAccess) List(ctx context.Context, prefix string) ([]string, error) {
	return p.physical.List(ctx, prefix)
}

func (p *PAccess) Purge(ctx context.Context) {
	if purgeable, ok := p.physical.(ToggleablePurgemonster); ok {
		purgeable.Purge(ctx)
	}
}
