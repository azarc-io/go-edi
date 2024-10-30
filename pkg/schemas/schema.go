package schemas

import (
	_ "embed"
	"encoding/json"
	"github.com/azarc-io/go-edi/internal/model"
)

//go:embed edi/edi-schema_contrl_v1.json
var SchemaContrl_v1 []byte

//go:embed edi/edi-schema_aperak_v1.json
var SchemaAperak_v1 []byte

//go:embed edi/edi-schema_cuscar-fri_v1.json
var SchemaCuscarFri_v1 []byte

//go:embed edi/edi-schema_cuscar-frc_v1.json
var SchemaCuscarFrc_v1 []byte

// LoadSchema loads the JSON Schema from a file and parses it into a Schema
func LoadSchema(data []byte) (*model.Schema, error) {
	var schema model.Schema
	err := json.Unmarshal(data, &schema)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}
