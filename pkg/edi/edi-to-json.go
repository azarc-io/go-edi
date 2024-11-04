package edi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	edi_reader "github.com/azarc-io/go-edi/internal/edi-reader"
	"github.com/azarc-io/go-edi/internal/model"
	"io"
)

const (
	SystemMaxItemCount = 10000
)

var (
	ErrEdiComponentElementIncorrectLocation = errors.New("edi component or element must be within a segment")
	ErrArrayRequiresItem                    = errors.New("array's must contain the item property")
	ErrSystemMaxItemCountExceeded           = errors.New("maximum array item count exceeded")
)

// Unmarshal parses an EDI string based on the loaded JSON Schema and returns a map
func Unmarshal(schema *model.Schema, input []byte, out any) error {
	reader := edi_reader.NewEDIReader(input)
	var buffer bytes.Buffer

	if schema.Type == model.JsonSchemaTypeUnknown {
		schema.Type = model.JsonSchemaTypeObject
	}

	if _, err := processProperty(schema.Property, reader, &buffer); err != nil {
		return err
	}

	data := buffer.Bytes()
	if len(data) != 0 {
		// inject the $schema
		if data[0] == byte('{') && schema.Self != nil && schema.Self.Schema != "" {
			data = append([]byte(fmt.Sprintf(`{"$schema": "%s", `, schema.Self.Schema)), data[1:]...)
		}
		return json.Unmarshal(data, &out)
	}
	return nil
}

func processProperty(p *model.Property, reader *edi_reader.EDIReader, out io.Writer) (bool, error) {
	switch p.Type {
	case model.JsonSchemaTypeArray:
		if p.Items == nil {
			return false, ErrArrayRequiresItem
		}
		_, _ = out.Write([]byte("["))
		var count int
		for {
			var childOut bytes.Buffer
			if added, err := processProperty(p.Items, reader, &childOut); err != nil {
				return false, err
			} else if added {
				if count > 0 {
					_, _ = out.Write([]byte(", "))
				}
				out.Write(childOut.Bytes())
				count++

				// check max value
				if p.MaxItems != 0 && p.MaxItems <= count {
					// at max items, break
					break
				}

				if SystemMaxItemCount <= count {
					return false, fmt.Errorf("%w: maxItems on an array must be set and be less than %d: got %d", ErrSystemMaxItemCountExceeded, SystemMaxItemCount, count)
				}
			} else {
				break
			}
		}
		_, _ = out.Write([]byte("]"))
		return count > 0, nil
	case model.JsonSchemaTypeObject:
		switch p.XEdi.Type {
		case model.EdiTypeSegment:
			return processSegment(p, reader, out)
		case model.EdiTypeComponent, model.EdiTypeElement:
			return false, fmt.Errorf("%w: type %s", ErrEdiComponentElementIncorrectLocation, p.XEdi.Type)
		default:
			// this is normal object, try process
			return processObject(p, reader, out)
		}
	default:
		if p.Const != nil {
			d, _ := json.Marshal(p.Const)
			_, _ = out.Write(d)
			return true, nil
		}
		return false, nil
	}
}

func processObject(property *model.Property, reader *edi_reader.EDIReader, out io.Writer) (bool, error) {
	_, _ = out.Write([]byte("{"))
	children := getOrderedEdi(property.Properties)
	var count int
	for _, child := range children {
		var childOut bytes.Buffer
		added, err := processProperty(child.Property, reader, &childOut)
		if err != nil {
			return false, err
		}
		if added {
			if count > 0 {
				_, _ = out.Write([]byte(`, `))
			}
			_, _ = out.Write([]byte(fmt.Sprintf(`"%s": %s`, child.Name, childOut.String())))
			count++
		}
	}
	_, _ = out.Write([]byte("}"))
	return count > 0, nil
}

func processSegment(property *model.Property, reader *edi_reader.EDIReader, out io.Writer) (bool, error) {
	seg, err := reader.PeekSegment()
	if err != nil {
		return false, err
	}
	if seg == "" {
		return false, nil
	}
	// ensure this is the correct segment
	if seg != property.XEdi.Tag && property.XEdi.Tag != "" {
		return false, nil
	}

	segment, err := reader.ReadSegment()
	if err != nil {
		return false, err
	}

	var segmentOut bytes.Buffer
	children := getOrderedEdi(property.Properties)
	components := segment.ReadComponents()
	var totalChildren int
	var componentIndex int
	for _, child := range children {
		var childOut bytes.Buffer
		var added bool
		switch child.Property.XEdi.Type {
		case model.EdiTypeComponent:
			if len(components) > componentIndex {
				comp := components[componentIndex]
				if err := addComponent(child.Property, comp, &childOut); err != nil {
					return false, err
				}
				componentIndex++
				added = true
			}
		case model.EdiTypeElement:
			if len(components) > componentIndex {
				comp := components[componentIndex]
				elms := comp.Elements()
				if len(elms) > 0 && elms[0] != "" {
					if err := addElement(child.Property, elms[0], &childOut); err != nil {
						return false, err
					}
				}
				componentIndex++
				added = true
			}
		default:
			if added, err = processProperty(child.Property, reader, &childOut); err != nil {
				return false, err
			}
		}

		if added && childOut.Len() > 0 {
			if totalChildren > 0 {
				_, _ = segmentOut.Write([]byte(`, `))
			}
			_, _ = segmentOut.Write([]byte(fmt.Sprintf(`"%s": %s`, child.Name, childOut.String())))
			totalChildren++
		}
	}

	if segmentOut.Len() > 0 {
		_, _ = out.Write([]byte("{"))
		_, _ = out.Write(segmentOut.Bytes())
		_, _ = out.Write([]byte("}"))
		return true, nil
	}
	return false, nil
}

func addComponent(property *model.Property, comp *edi_reader.Component, out io.Writer) error {
	var compOut bytes.Buffer
	elements := comp.Elements()
	children := getOrderedEdi(property.Properties)
	var totalChildren int
	for i, child := range children {
		if len(elements) <= i {
			break
		}
		if elements[i] == "" {
			continue
		}

		if totalChildren > 0 {
			_, _ = compOut.Write([]byte(`, `))
		}
		totalChildren++
		_, _ = compOut.Write([]byte(fmt.Sprintf(`"%s": `, child.Name)))
		if err := addElement(child.Property, elements[i], &compOut); err != nil {
			return err
		}
	}
	if compOut.Len() > 0 {
		_, _ = out.Write([]byte("{"))
		_, _ = out.Write(compOut.Bytes())
		_, _ = out.Write([]byte("}"))
	}
	return nil
}

func addElement(element *model.Property, value string, out io.Writer) error {
	switch element.Type {
	case model.JsonSchemaTypeString:
		d, _ := json.Marshal(value)
		_, _ = out.Write(d)
	default:
		_, _ = out.Write([]byte(value))
	}
	return nil
}
