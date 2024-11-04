package model

type EdiType string
type JsonSchemaType string

const (
	EdiTypeSegment   EdiType = "segment"
	EdiTypeElement   EdiType = "element"
	EdiTypeComponent EdiType = "component"

	JsonSchemaTypeArray   JsonSchemaType = "array"
	JsonSchemaTypeObject  JsonSchemaType = "object"
	JsonSchemaTypeNumber  JsonSchemaType = "number"
	JsonSchemaTypeString  JsonSchemaType = "string"
	JsonSchemaTypeBoolean JsonSchemaType = "boolean"
	JsonSchemaTypeInteger JsonSchemaType = "integer"
	JsonSchemaTypeUnknown JsonSchemaType = ""
)

type (
	Schema struct {
		Schema string `json:"$schema"`
		Self   *struct {
			Schema string `json:"$schema"`
		} `json:"self"`
		*Property
	}

	Property struct {
		Description string         `json:"description"`
		Type        JsonSchemaType `json:"type"`
		XEdi        XEdi           `json:"x-edi"`
		Properties  Properties     `json:"properties"`
		Items       *Property      `json:"items"`
		IsRequired  bool           `json:"isRequired"`
		Const       any            `json:"const"`
		MinItems    int            `json:"minItems"`
		MaxItems    int            `json:"maxItems"`
	}

	Properties map[string]*Property

	XEdi struct {
		Type  EdiType `json:"type"`
		Order int     `json:"order"`
		Tag   string  `json:"tag"`
	}
)
