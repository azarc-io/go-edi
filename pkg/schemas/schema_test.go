package schemas_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	edi "github.com/azarc-io/go-edi/pkg/edi"
	"github.com/azarc-io/go-edi/pkg/schemas"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

//go:embed edi/edi-schema_contrl_v1.json
var SchemaContrl_v1 []byte

//go:embed edi/edi-schema_aperak_v1.json
var SchemaAperak_v1 []byte

//go:embed edi/edi-schema_cuscar-fri_v1.json
var SchemaCuscarFri_v1 []byte

//go:embed edi/edi-schema_cuscar-frc_v1.json
var SchemaCuscarFrc_v1 []byte

//go:embed fixtures
var ff embed.FS

func TestSchemas(t *testing.T) {
	tests := []struct {
		Name   string
		Schema []byte
	}{
		{Name: "contrl", Schema: SchemaContrl_v1},
		{Name: "contrl-error", Schema: SchemaContrl_v1},
		{Name: "aperak", Schema: SchemaAperak_v1},
		{Name: "cuscar-fri", Schema: SchemaCuscarFri_v1},
		{Name: "cuscar-frc", Schema: SchemaCuscarFrc_v1},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Marshal: %s", test.Name), func(t *testing.T) {
			g := goldie.New(
				t,
				goldie.WithFixtureDir("fixtures"),
				goldie.WithNameSuffix(".golden.json"),
			)
			input, err := ff.ReadFile(fmt.Sprintf("fixtures/%s.golden.edi", test.Name))
			require.NoError(t, err)
			s, err := schemas.LoadSchema(test.Schema)
			require.NoError(t, err)
			m := make(map[string]any)
			err = edi.Unmarshal(s, input, &m)
			require.NoError(t, err)
			g.AssertJson(t, test.Name, m)
		})

		t.Run(fmt.Sprintf("Unmarshal: %s", test.Name), func(t *testing.T) {
			g := goldie.New(
				t,
				goldie.WithFixtureDir("fixtures"),
				goldie.WithNameSuffix(".golden.edi"),
			)
			input, err := ff.ReadFile(fmt.Sprintf("fixtures/%s.golden.json", test.Name))
			require.NoError(t, err)
			s, err := schemas.LoadSchema(test.Schema)
			require.NoError(t, err)

			m := make(map[string]any)
			require.NoError(t, json.Unmarshal(input, &m))
			data, err := edi.Marshal(s, &m, edi.WithSegmentSeparator("'\n"))
			require.NoError(t, err)
			g.Assert(t, test.Name, data)
		})
	}

	t.Run("broken schema", func(t *testing.T) {
		s, err := schemas.LoadSchema([]byte(`{"broken": ...}`))
		assert.Nil(t, s)
		assert.Error(t, err)
	})
}
