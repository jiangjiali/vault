package logical

import (
	"context"

	"github.com/jiangjiali/vault/sdk/physical"
)

type LStorage struct {
	underlying physical.Backend
}

func (s *LStorage) Get(ctx context.Context, key string) (*StorageEntry, error) {
	entry, err := s.underlying.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	return &StorageEntry{
		Key:      entry.Key,
		Value:    entry.Value,
		SealWrap: entry.SealWrap,
	}, nil
}

func (s *LStorage) Put(ctx context.Context, entry *StorageEntry) error {
	return s.underlying.Put(ctx, &physical.Entry{
		Key:      entry.Key,
		Value:    entry.Value,
		SealWrap: entry.SealWrap,
	})
}

func (s *LStorage) Delete(ctx context.Context, key string) error {
	return s.underlying.Delete(ctx, key)
}

func (s *LStorage) List(ctx context.Context, prefix string) ([]string, error) {
	return s.underlying.List(ctx, prefix)
}

func (s *LStorage) Underlying() physical.Backend {
	return s.underlying
}

func NewLStorage(underlying physical.Backend) *LStorage {
	return &LStorage{
		underlying: underlying,
	}
}
