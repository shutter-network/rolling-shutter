package db

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	sqlcFileName     = "sqlc.yaml"
	ErrNotADirectory = errors.New("not a directory")
)

type data struct {
	SQL []entry `yaml:"sql"`
}
type entry struct {
	Schema string
	Gen    struct {
		Golang struct{ Out string } `yaml:"go"`
	}
}

// ParseSQLC parses the 'sqlc.yaml' file and reads in the
// paths of all the schema definition directories.
// The schema file will then be stored in the object's state
// together with the passed in `version`.
func ParseSQLC(filesystem fs.FS, sqlcPath string, version int) ([]Schema, error) {
	b, err := fs.ReadFile(filesystem, sqlcPath)
	if err != nil {
		return nil, err
	}
	sqlcDir := path.Dir(sqlcPath)

	d := &data{}
	err = yaml.Unmarshal(b, &d)
	if err != nil {
		return nil, err
	}

	schemas := []Schema{}
	for _, entry := range d.SQL {
		// TODO: allow for individual sql files and multiple dirs.
		// For now assume we only get one string
		// which is the path of the directory containing the schema files.
		// This won't handle multiple directories, or individual sql file names.
		schemaDirPath := path.Join(sqlcDir, entry.Schema)
		dirEntries, err := fs.ReadDir(filesystem, schemaDirPath)
		if err != nil {
			return nil, errors.Wrap(err, "error reading schema dir")
		}
		for _, de := range dirEntries {
			i, err := de.Info()
			if err != nil {
				return nil, err
			}
			if i.IsDir() {
				continue
			}
			// TODO: allow database migrations.
			// For now assume only schemas, no migrations.
			// However sqlc does recognize migration files from different tools.
			base, isSQL := strings.CutSuffix(i.Name(), ".sql")
			if !isSQL {
				continue
			}
			pathstr := path.Join(schemaDirPath, i.Name())
			if version != 1 {
				pathstr = path.Join(schemaDirPath, fmt.Sprintf("v%d", version), i.Name())
			}

			schema := Schema{
				Version: version,
				Name:    base,
				Path:    pathstr,
			}
			schemas = append(schemas, schema)
		}
	}
	return schemas, nil
}

func NewSQLCDefinition(filesystem fs.FS, sqlcPath string, name string, version int) (*SQLC, error) {
	p := path.Clean(sqlcPath)
	des, err := fs.ReadDir(filesystem, p)
	if errors.Is(err, ErrNotADirectory) {
		p = path.Base(p)
	} else if err != nil {
		return nil, err
	}
	p = path.Join(p, sqlcFileName)
	var foundPath string
	for _, d := range des {
		if d.Name() == sqlcFileName && !d.IsDir() {
			foundPath = p
			break
		}
	}
	if foundPath == "" {
		return nil, errors.Errorf("SQLC file '%s' does not exists", p)
	}
	schemas, err := ParseSQLC(filesystem, foundPath, version)
	if err != nil {
		return nil, err
	}
	print("name", name)
	return &SQLC{
		schemas:    schemas,
		filesystem: filesystem,
		name:       name,
		sqlcPath:   sqlcPath,
	}, nil
}

// SQLC implements the `Definition` interface and keeps
// information about the database schemas.
type SQLC struct {
	schemas    []Schema
	filesystem fs.FS
	name       string
	sqlcPath   string
}

func (d *SQLC) Name() string {
	return d.name
}

// Init is a NOOP for the SQLCDatabase.
func (d *SQLC) Init(_ context.Context, _ pgx.Tx) error {
	return nil
}

// Create reads in the content of all memorized schema definition
// `.sql` files and executes it's create statements.
// Also, all schema versions are written in the database`s `meta`
// table for later validation.
func (d *SQLC) Create(ctx context.Context, tx pgx.Tx) error {
	for _, s := range d.sqlCreateStatements() {
		_, err := tx.Exec(ctx, s)
		if err != nil {
			return errors.Wrapf(err, "failed to execute SQL statements for definition '%s'", d.Name())
		}
	}

	for _, schema := range d.schemas {
		// this is initial creation of db, so create version as one here
		schema.Version = 1
		err := InsertSchemaVersion(ctx, tx, d.Name(), schema)
		if err != nil {
			return err
		}
	}
	return nil
}

// Validate compares the schema versions of the connected to database
// with the schema versions of it's schema definitions.
func (d *SQLC) Validate(ctx context.Context, tx pgx.Tx) error {
	for _, schema := range d.schemas {
		err := ValidateSchemaVersion(ctx, tx, d.Name(), schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *SQLC) sqlCreateStatements() []string {
	sqlStatements := []string{}
	for _, schema := range d.schemas {
		b, err := fs.ReadFile(d.filesystem, schema.Path)
		if err != nil {
			panic(err)
		}
		sqlStatements = append(sqlStatements, string(b))
	}
	return sqlStatements
}

func (d *SQLC) LoadMigrations() ([]Migration, error) {
	migrationsPath := path.Join(path.Dir(d.sqlcPath), "migrations")
	entries, err := fs.ReadDir(d.filesystem, migrationsPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error reading migrations directory")
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		var fileversion int
		_, err := fmt.Sscanf(name, "V%d_", &fileversion)
		if err != nil {
			continue
		}

		_, err = fs.ReadFile(d.filesystem, path.Join(migrationsPath, name))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read migration %s", name)
		}

		migrations = append(migrations, Migration{
			Version: fileversion,
			Path:    path.Join(migrationsPath, name),
			Up:      true,
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (d *SQLC) Migrate(ctx context.Context, tx pgx.Tx) error {
	for _, schema := range d.schemas {
		// Get current version from meta-inf
		version, err := GetSchemaVersion(ctx, tx, d.Name(), schema)
		if err != nil {
			return err
		}

		migrations, err := d.LoadMigrations()
		if err != nil {
			return errors.Wrap(err, "failed to load migrations")
		}

		// Apply only migrations that are newer than current version
		for _, migration := range migrations {
			println(migration.Path, migration.Version, "migration")
			if migration.Version <= version {
				continue
			}

			content, err := fs.ReadFile(d.filesystem, migration.Path)
			if err != nil {
				return errors.Wrapf(err, "failed to read migration file %s", migration.Path)
			}

			log.Info().
				Str("definition", d.Name()).
				Int("from_version", version).
				Int("to_version", migration.Version).
				Str("path", migration.Path).
				Msg("applying migration")

			_, err = tx.Exec(ctx, string(content))
			if err != nil {
				return errors.Wrapf(err, "failed to apply migration %d", migration.Version)
			}

			// Update version after each successful migration
			err = UpdateSchemaVersion(ctx, tx, d.Name(), schema)
			if err != nil {
				return errors.Wrapf(err, "failed to update schema version to %d", migration.Version)
			}
		}
	}

	return nil
}
