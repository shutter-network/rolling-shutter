package db

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// InitDB initializes an empty database with the all schema definitions as specified from
// the passed in `Definition`'s Create() and Init() actions.
// Additionally, a `role` will be written in the database's meta key/value store, that
// pins the database to a specific role, e.g. "keyper-test" or "snapshot-keyper-production"
// in order to prevent later usage of the database with commands that fulfill a different role.
func InitDB(ctx context.Context, dbpool *pgxpool.Pool, role string, definition Definition) error {
	// First check if schema exists and is valid
	err := dbpool.BeginFunc(WrapContext(ctx, definition.Validate))
	if err == nil {
		shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database already exists")
		return nil
	} else if errors.Is(err, ErrNeedsMigation) {
		// Schema exists, just run migrations
		shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database exists, checking for migrations")
		err = dbpool.BeginFunc(WrapContext(ctx, definition.Migrate))
		if err != nil {
			return errors.Wrap(err, "failed to apply migrations")
		}
		return nil
	} else if !errors.Is(err, ErrValueMismatch) && !errors.Is(err, ErrKeyNotFound) {
		return err
	}
	// Schema doesn't exist or is invalid, create it
	err = dbpool.BeginFunc(WrapContext(ctx, definition.Create))
	if err != nil {
		return err
	}

	err = dbpool.BeginFunc(WrapContext(ctx, definition.Init))
	if err != nil {
		return err
	}

	// Run any migrations after initial creation
	err = dbpool.BeginFunc(WrapContext(ctx, definition.Migrate))
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}

	// Set the database role
	err = dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return InsertDBVersion(ctx, tx, role)
	})
	if err != nil {
		return err
	}

	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database initialized")
	return nil
}

var _ Definition = AggregateDefinition{}

// NewAggregateDefinition constructs a new AggregateDefinition instance
// and implements the `Definition` interface.
// The passed in `definitions` will be stored in the AggregateDefinition's
// state and a call to the AggregateDefinition's methods will be dispatched to all
// stored definitions.
// Importantly, this constructor will unpack the internal definitions of any passed in,
// wrapped AggregateDefinition. This means that the outer AggregateDefinition always stores
// the flattened set of all underlying child-definitions.
func NewAggregateDefinition(name string, definitions ...Definition) AggregateDefinition {
	defs := map[Definition]bool{}
	for _, def := range definitions {
		nestedDefs, isWrapper := def.(AggregateDefinition)
		if isWrapper {
			for deff := range nestedDefs.defs {
				defs[deff] = true
			}
		} else {
			defs[def] = true
		}
	}
	return AggregateDefinition{name: name, defs: defs}
}

type AggregateDefinition struct {
	name string
	defs map[Definition]bool
}

func (d AggregateDefinition) Name() string {
	return d.name
}

func (d AggregateDefinition) Init(ctx context.Context, tx pgx.Tx) error {
	for def := range d.defs {
		err := def.Init(ctx, tx)
		if err != nil {
			return errors.Wrapf(err, "can't initialize DB for definition '%s'", def.Name())
		}
	}
	return nil
}

func (d AggregateDefinition) Create(ctx context.Context, tx pgx.Tx) error {
	err := metaDefinition.Create(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "can't create DB for meta definition")
	}
	for def := range d.defs {
		err := def.Create(ctx, tx)
		if err != nil {
			return errors.Wrapf(err, "can't create DB for definition '%s'", def.Name())
		}
	}
	return nil
}

func (d AggregateDefinition) Validate(ctx context.Context, tx pgx.Tx) error {
	for def := range d.defs {
		err := def.Validate(ctx, tx)
		if err != nil {
			return errors.Wrapf(err, "validation error for DB '%s'", def.Name())
		}
	}
	return nil
}

func (d AggregateDefinition) Migrate(ctx context.Context, tx pgx.Tx) error {
	for def := range d.defs {
		err := def.Migrate(ctx, tx)
		if err != nil {
			return errors.Wrapf(err, "migration failed for definition '%s'", def.Name())
		}
	}
	return nil
}

type Definition interface {
	Name() string
	Create(context.Context, pgx.Tx) error
	Init(context.Context, pgx.Tx) error
	Validate(context.Context, pgx.Tx) error
	Migrate(context.Context, pgx.Tx) error
}

type Schema struct {
	Version int
	Name    string
	Path    string
}

type Migration struct {
	Version int
	Path    string
	Up      bool // up or down migration
}
