package model

type EdiType string

const (
	EdiTypeSegment   EdiType = "segment"
	EdiTypeElement   EdiType = "element"
	EdiTypeComponent EdiType = "component"
)

type (
	Schema struct {
		Schema string `json:"$schema"`
		*Property
	}

	Property struct {
		Description string     `json:"description"`
		Type        string     `json:"type"`
		XEdi        XEdi       `json:"x-edi"`
		Properties  Properties `json:"properties"`
		Items       *Property  `json:"items"`
		IsRequired  bool       `json:"isRequired"`
		Const       any        `json:"const"`
		MinItems    int        `json:"minItems"`
		MaxItems    int        `json:"maxItems"`
	}

	Properties map[string]*Property

	XEdi struct {
		Type  EdiType `json:"type"`
		Order int     `json:"order"`
		Tag   string  `json:"tag"`
	}
)
