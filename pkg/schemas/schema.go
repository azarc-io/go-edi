package schemas

import (
	_ "embed"
	"encoding/json"
	"github.com/azarc-io/go-edi/internal/model"
)

// LoadSchema loads the JSON Schema from a file and parses it into a Schema
func LoadSchema(data []byte) (*model.Schema, error) {
	var schema model.Schema
	err := json.Unmarshal(data, &schema)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}
