package snapshot

import (
	"io"
	"text/template"

	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/medley"
)

type Config struct {
	DatabaseURL string
}

const configTemplate = `# Shutter snapshot config
`

var tmpl *template.Template = medley.MustBuildTemplate("snapshot", configTemplate)

func (config Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}

// Unmarshal unmarshals a SnapshotConfig from the given Viper object.
func (config *Config) Unmarshal(v *viper.Viper) error {
	return nil
}
