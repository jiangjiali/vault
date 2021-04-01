package inmem

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	log "github.com/jiangjiali/vault/sdk/helper/hclutil/hclog"
	"github.com/jiangjiali/vault/sdk/physical"

	"github.com/jiangjiali/vault/sdk/helper/armon/radix"
)

// Verify interfaces are satisfied
var _ physical.Backend = (*IBackend)(nil)
var _ physical.Transactional = (*TransactionalInmemBackend)(nil)

var (
	PutDisabledError    = errors.New("put operations disabled in inmem backend")
	GetDisabledError    = errors.New("get operations disabled in inmem backend")
	DeleteDisabledError = errors.New("delete operations disabled in inmem backend")
	ListDisabledError   = errors.New("list operations disabled in inmem backend")
)

// InmemBackend is an in-memory only physical backend. It is useful
// for testing and development situations where the data is not
// expected to be durable.
type IBackend struct {
	sync.RWMutex
	root       *radix.Tree
	permitPool *physical.PermitPool
	logger     log.Logger
	failGet    *uint32
	failPut    *uint32
	failDelete *uint32
	failList   *uint32
	logOps     bool
}

type TransactionalInmemBackend struct {
	IBackend
}

// NewInmem constructs a new in-memory backend
func NewInmem(_ map[string]string, logger log.Logger) (physical.Backend, error) {
	in := &IBackend{
		root:       radix.New(),
		permitPool: physical.NewPermitPool(physical.DefaultParallelOperations),
		logger:     logger,
		failGet:    new(uint32),
		failPut:    new(uint32),
		failDelete: new(uint32),
		failList:   new(uint32),
		logOps:     os.Getenv("VAULT_INMEM_LOG_ALL_OPS") != "",
	}
	return in, nil
}

// Basically for now just creates a permit pool of size 1 so only one operation
// can run at a time
func NewTransactionalInmem(_ map[string]string, logger log.Logger) (physical.Backend, error) {
	in := &TransactionalInmemBackend{
		IBackend: IBackend{
			root:       radix.New(),
			permitPool: physical.NewPermitPool(1),
			logger:     logger,
			failGet:    new(uint32),
			failPut:    new(uint32),
			failDelete: new(uint32),
			failList:   new(uint32),
			logOps:     os.Getenv("VAULT_INMEM_LOG_ALL_OPS") != "",
		},
	}
	return in, nil
}

// Put is used to insert or update an entry
func (i *IBackend) Put(ctx context.Context, entry *physical.Entry) error {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.Lock()
	defer i.Unlock()

	return i.PutInternal(ctx, entry)
}

func (i *IBackend) PutInternal(ctx context.Context, entry *physical.Entry) error {
	if i.logOps {
		i.logger.Trace("put", "key", entry.Key)
	}
	if atomic.LoadUint32(i.failPut) != 0 {
		return PutDisabledError
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	i.root.Insert(entry.Key, entry.Value)
	return nil
}

func (i *IBackend) FailPut(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failPut, val)
}

// Get is used to fetch an entry
func (i *IBackend) Get(ctx context.Context, key string) (*physical.Entry, error) {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.RLock()
	defer i.RUnlock()

	return i.GetInternal(ctx, key)
}

func (i *IBackend) GetInternal(ctx context.Context, key string) (*physical.Entry, error) {
	if i.logOps {
		i.logger.Trace("get", "key", key)
	}
	if atomic.LoadUint32(i.failGet) != 0 {
		return nil, GetDisabledError
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if raw, ok := i.root.Get(key); ok {
		return &physical.Entry{
			Key:   key,
			Value: raw.([]byte),
		}, nil
	}
	return nil, nil
}

func (i *IBackend) FailGet(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failGet, val)
}

// Delete is used to permanently delete an entry
func (i *IBackend) Delete(ctx context.Context, key string) error {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.Lock()
	defer i.Unlock()

	return i.DeleteInternal(ctx, key)
}

func (i *IBackend) DeleteInternal(ctx context.Context, key string) error {
	if i.logOps {
		i.logger.Trace("delete", "key", key)
	}
	if atomic.LoadUint32(i.failDelete) != 0 {
		return DeleteDisabledError
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	i.root.Delete(key)
	return nil
}

func (i *IBackend) FailDelete(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failDelete, val)
}

// List is used to list all the keys under a given
// prefix, up to the next prefix.
func (i *IBackend) List(ctx context.Context, prefix string) ([]string, error) {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.RLock()
	defer i.RUnlock()

	return i.ListInternal(ctx, prefix)
}

func (i *IBackend) ListInternal(ctx context.Context, prefix string) ([]string, error) {
	if i.logOps {
		i.logger.Trace("list", "prefix", prefix)
	}
	if atomic.LoadUint32(i.failList) != 0 {
		return nil, ListDisabledError
	}

	var out []string
	seen := make(map[string]interface{})
	walkFn := func(s string, v interface{}) bool {
		trimmed := strings.TrimPrefix(s, prefix)
		sep := strings.Index(trimmed, "/")
		if sep == -1 {
			out = append(out, trimmed)
		} else {
			trimmed = trimmed[:sep+1]
			if _, ok := seen[trimmed]; !ok {
				out = append(out, trimmed)
				seen[trimmed] = struct{}{}
			}
		}
		return false
	}
	i.root.WalkPrefix(prefix, walkFn)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return out, nil
}

func (i *IBackend) FailList(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failList, val)
}

// Implements the transaction interface
func (t *TransactionalInmemBackend) Transaction(ctx context.Context, txns []*physical.TxnEntry) error {
	t.permitPool.Acquire()
	defer t.permitPool.Release()

	t.Lock()
	defer t.Unlock()

	return physical.GenericTransactionHandler(ctx, t, txns)
}
