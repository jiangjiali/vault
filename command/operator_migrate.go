package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jiangjiali/vault/command/server"
	"github.com/jiangjiali/vault/sdk/helper/complete"
	"github.com/jiangjiali/vault/sdk/helper/errors"
	"github.com/jiangjiali/vault/sdk/helper/errwrap"
	"github.com/jiangjiali/vault/sdk/helper/hclutil/hcl"
	"github.com/jiangjiali/vault/sdk/helper/hclutil/hcl/hcl/ast"
	log "github.com/jiangjiali/vault/sdk/helper/hclutil/hclog"
	"github.com/jiangjiali/vault/sdk/helper/logging"
	"github.com/jiangjiali/vault/sdk/helper/mitchellh/cli"
	"github.com/jiangjiali/vault/sdk/physical"
	"github.com/jiangjiali/vault/vault"
)

var _ cli.Command = (*OperatorMigrateCommand)(nil)
var _ cli.CommandAutocomplete = (*OperatorMigrateCommand)(nil)

var errAbort = errors.New("Migration aborted")

type OperatorMigrateCommand struct {
	*BaseCommand

	PhysicalBackends map[string]physical.Factory
	flagConfig       string
	flagStart        string
	flagReset        bool
	logger           log.Logger
	ShutdownCh       chan struct{}
}

type migratorConfig struct {
	StorageSource      *server.Storage `hcl:"-"`
	StorageDestination *server.Storage `hcl:"-"`
}

func (c *OperatorMigrateCommand) Synopsis() string {
	return "在存储后端之间迁移保安全存储数据"
}

func (c *OperatorMigrateCommand) Help() string {
	helpText := `
使用: vault operator migrate [选项]

  此命令启动存储后端迁移过程，以将所有数据从一个后端复制到另一个后端。
  这直接对加密数据进行操作，不需要安全库服务器，也不需要任何解封操作。

  使用配置文件启动迁移：

      $ vault operator migrate -config=migrate.hcl

  有关更多信息，请参阅文档。

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorMigrateCommand) Flags() *FlagSets {
	set := NewFlagSets()
	f := set.NewFlagSet("命令选项")

	f.StringVar(&StringVar{
		Name:   "config",
		Target: &c.flagConfig,
		Completion: complete.PredictOr(
			complete.PredictFiles("*.hcl"),
		),
		Usage: "配置文件的路径。此配置文件应仅包含migrator指令。",
	})

	f.StringVar(&StringVar{
		Name:   "start",
		Target: &c.flagStart,
		Usage:  "仅在该值处或之后按字典顺序复制键。",
	})

	f.BoolVar(&BoolVar{
		Name:   "reset",
		Target: &c.flagReset,
		Usage:  "重置迁移锁定。不会发生迁移。",
	})

	return set
}

func (c *OperatorMigrateCommand) AutocompleteArgs() complete.Predictor {
	return nil
}

func (c *OperatorMigrateCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *OperatorMigrateCommand) Run(args []string) int {
	c.logger = logging.NewVaultLogger(log.Info)
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if c.flagConfig == "" {
		c.UI.Error("Must specify exactly one config path using -config")
		return 1
	}

	config, err := c.loadMigratorConfig(c.flagConfig)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error loading configuration from %s: %s", c.flagConfig, err))
		return 1
	}

	if err := c.migrate(config); err != nil {
		if err == errAbort {
			return 0
		}
		c.UI.Error(fmt.Sprintf("Error migrating: %s", err))
		return 2
	}

	if c.flagReset {
		c.UI.Output("Success! Migration lock reset (if it was set).")
	} else {
		c.UI.Output("Success! All of the keys have been migrated.")
	}

	return 0
}

// migrate attempts to instantiate the source and destinations backends,
// and then invoke the migration the the root of the keyspace.
func (c *OperatorMigrateCommand) migrate(config *migratorConfig) error {
	from, err := c.newBackend(config.StorageSource.Type, config.StorageSource.Config)
	if err != nil {
		return errwrap.Wrapf("error mounting 'storage_source': {{err}}", err)
	}

	if c.flagReset {
		if err := SetStorageMigration(from, false); err != nil {
			return errwrap.Wrapf("error reseting migration lock: {{err}}", err)
		}
		return nil
	}

	to, err := c.newBackend(config.StorageDestination.Type, config.StorageDestination.Config)
	if err != nil {
		return errwrap.Wrapf("error mounting 'storage_destination': {{err}}", err)
	}

	migrationStatus, err := CheckStorageMigration(from)
	if err != nil {
		return errwrap.Wrapf("error checking migration status: {{err}}", err)
	}

	if migrationStatus != nil {
		return fmt.Errorf("storage migration in progress (started: %s)", migrationStatus.Start.Format(time.RFC3339))
	}

	if err := SetStorageMigration(from, true); err != nil {
		return errwrap.Wrapf("error setting migration lock: {{err}}", err)
	}

	defer SetStorageMigration(from, false)

	ctx, cancelFunc := context.WithCancel(context.Background())

	doneCh := make(chan error)
	go func() {
		doneCh <- c.migrateAll(ctx, from, to)
	}()

	select {
	case err := <-doneCh:
		cancelFunc()
		return err
	case <-c.ShutdownCh:
		c.UI.Output("==> Migration shutdown triggered\n")
		cancelFunc()
		<-doneCh
		return errAbort
	}
}

// migrateAll copies all keys in lexicographic order.
func (c *OperatorMigrateCommand) migrateAll(ctx context.Context, from physical.Backend, to physical.Backend) error {
	return dfsScan(ctx, from, func(ctx context.Context, path string) error {
		if path < c.flagStart || path == storageMigrationLock || path == vault.CoreLockPath {
			return nil
		}

		entry, err := from.Get(ctx, path)

		if err != nil {
			return errwrap.Wrapf("error reading entry: {{err}}", err)
		}

		if entry == nil {
			return nil
		}

		if err := to.Put(ctx, entry); err != nil {
			return errwrap.Wrapf("error writing entry: {{err}}", err)
		}
		c.logger.Info("copied key", "path", path)
		return nil
	})
}

func (c *OperatorMigrateCommand) newBackend(kind string, conf map[string]string) (physical.Backend, error) {
	factory, ok := c.PhysicalBackends[kind]
	if !ok {
		return nil, fmt.Errorf("no Vault storage backend named: %+q", kind)
	}

	return factory(conf, c.logger)
}

// loadMigratorConfig loads the configuration at the given path
func (c *OperatorMigrateCommand) loadMigratorConfig(path string) (*migratorConfig, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("location is a directory, not a file")
	}

	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	obj, err := hcl.ParseBytes(d)
	if err != nil {
		return nil, err
	}

	var result migratorConfig
	if err := hcl.DecodeObject(&result, obj); err != nil {
		return nil, err
	}

	list, ok := obj.Node.(*ast.ObjectList)
	if !ok {
		return nil, fmt.Errorf("error parsing: file doesn't contain a root object")
	}

	// Look for storage_* stanzas
	for _, stanza := range []string{"storage_source", "storage_destination"} {
		o := list.Filter(stanza)
		if len(o.Items) != 1 {
			return nil, fmt.Errorf("exactly one '%s' block is required", stanza)
		}

		if err := parseStorage(&result, o, stanza); err != nil {
			return nil, errwrap.Wrapf("error parsing '%s': {{err}}", err)
		}
	}
	return &result, nil
}

// parseStorage reuses the existing storage parsing that's part of the main Vault
// config processing, but only keeps the storage result.
func parseStorage(result *migratorConfig, list *ast.ObjectList, name string) error {
	tmpConfig := new(server.Config)

	if err := server.ParseStorage(tmpConfig, list, name); err != nil {
		return err
	}

	switch name {
	case "storage_source":
		result.StorageSource = tmpConfig.Storage
	case "storage_destination":
		result.StorageDestination = tmpConfig.Storage
	default:
		return fmt.Errorf("unknown storage name: %s", name)
	}

	return nil
}

// dfsScan will invoke cb with every key from source.
// Keys will be traversed in lexicographic, depth-first order.
func dfsScan(ctx context.Context, source physical.Backend, cb func(ctx context.Context, path string) error) error {
	dfs := []string{""}

	for l := len(dfs); l > 0; l = len(dfs) {
		key := dfs[len(dfs)-1]
		if key == "" || strings.HasSuffix(key, "/") {
			children, err := source.List(ctx, key)
			if err != nil {
				return errwrap.Wrapf("failed to scan for children: {{err}}", err)
			}
			sort.Strings(children)

			// remove List-triggering key and add children in reverse order
			dfs = dfs[:len(dfs)-1]
			for i := len(children) - 1; i >= 0; i-- {
				if children[i] != "" {
					dfs = append(dfs, key+children[i])
				}
			}
		} else {
			err := cb(ctx, key)
			if err != nil {
				return err
			}

			dfs = dfs[:len(dfs)-1]
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
	return nil
}
