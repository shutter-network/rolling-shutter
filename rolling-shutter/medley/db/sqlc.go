package db

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
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
	println(sqlcPath, "sqlcpath")
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

			println(base, "base", pathstr, "pathstr")
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
	}, nil
}

// SQLC implements the `Definition` interface and keeps
// information about the database schemas.
type SQLC struct {
	schemas    []Schema
	filesystem fs.FS
	name       string
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
