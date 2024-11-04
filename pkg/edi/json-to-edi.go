package edi

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/azarc-io/go-edi/internal/model"
	"io"
	"regexp"
	"strings"
)

var (
	ErrUnexpectedInputType = errors.New("unexpected input type")
	ErrUnexpectedEDIType   = errors.New("unexpected edi type")
)

type Options struct {
	SegmentSeparator   string
	ComponentSeparator string
	ElementSeparator   string
	EscapeCharacter    string
	stringEscaper      func(string) string
}

type Option func(o *Options)

func WithSegmentSeparator(s string) Option {
	return func(o *Options) {
		o.SegmentSeparator = s
	}
}
func WithComponentSeparator(s string) Option {
	return func(o *Options) {
		o.ComponentSeparator = s
	}
}
func WithElementSeparator(s string) Option {
	return func(o *Options) {
		o.ElementSeparator = s
	}
}
func WithEscapeCharacter(e string) Option {
	return func(o *Options) {
		o.EscapeCharacter = e
	}
}

func Marshal(schema *model.Schema, input any, opts ...Option) ([]byte, error) {
	if !isMapStringAny(input) && !isStruct(input) {
		return nil, ErrUnexpectedInputType
	}

	toProcess, ok := input.(map[string]any)
	if !ok {
		r, _ := toMapRecursive(input)
		toProcess, _ = r.(map[string]any)
	}

	if schema.Type == model.JsonSchemaTypeUnknown {
		schema.Type = model.JsonSchemaTypeObject
	}

	o := &Options{
		SegmentSeparator:   model.DefaultSeparatorSegment,
		ComponentSeparator: model.DefaultSeparatorComponent,
		ElementSeparator:   model.DefaultSeparatorElement,
		EscapeCharacter:    model.DefaultEscapeCharacter,
	}

	for _, opt := range opts {
		opt(o)
	}

	var err error
	o.stringEscaper, err = createStringEscaper(o)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := writeEdi(schema.Property, toProcess, &out, o); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func writeEdi(p *model.Property, data any, out io.Writer, opts *Options) error {
	if p == nil || data == nil {
		return nil
	}

	switch p.Type {
	case model.JsonSchemaTypeObject:
		m, ok := data.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: expected data type to be map[string]any, got %T", ErrUnexpectedInputType, data)
		}
		items := getOrderedEdi(p.Properties)
		var processedSegment bool
		for _, item := range items {
			if item.Property.XEdi.Type == model.EdiTypeComponent || item.Property.XEdi.Type == model.EdiTypeElement {
				// this is the components or elements of a parent segment, process this outer, make sure segment not already been processed
				if processedSegment {
					continue
				}
				processedSegment = true
				if err := writeSegment(p, items, m, out, opts); err != nil {
					return err
				}
			} else {
				// this is a normal property which is not related to a EDI segment, recursively write
				if err := writeEdi(item.Property, m[item.Name], out, opts); err != nil {
					return err
				}
			}
		}
	case model.JsonSchemaTypeArray:
		if !isSlice(data) {
			return fmt.Errorf("%w: expected data type to be []any, got %T", ErrUnexpectedInputType, data)
		}
		a, ok := data.([]any)
		if !ok {
			return fmt.Errorf("%w: expected data type to be []map[string]any, got %t", ErrUnexpectedInputType, data)
		}
		for _, item := range a {
			if err := writeEdi(p.Items, item, out, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeSegment(p *model.Property, items []orderedEdiItem, data map[string]any, out io.Writer, opts *Options) error {
	if p.XEdi.Type != model.EdiTypeSegment {
		return fmt.Errorf("%w: expected 'segment', received '%s'", ErrUnexpectedEDIType, p.XEdi.Type)
	}
	_, err := out.Write([]byte(p.XEdi.Tag))
	if err != nil {
		return err
	}
	var compsBefore int
	for _, item := range items {
		switch item.Property.XEdi.Type {
		case model.EdiTypeComponent:
			var childOut bytes.Buffer
			if err := writeComponent(item.Property, data[item.Name], &childOut, opts); err != nil {
				return err
			} else if childOut.Len() > 0 {
				if _, err := out.Write([]byte(strings.Repeat(opts.ComponentSeparator, compsBefore+1))); err != nil {
					return err
				}
				compsBefore = 0
				if _, err := out.Write(childOut.Bytes()); err != nil {
					return err
				}
			} else {
				// didn't write anything for this component, increment counter
				compsBefore++
			}
		case model.EdiTypeElement:
			val := item.Property.Const
			if val == nil {
				val = data[item.Name]
			}
			if val != nil {
				vstr := opts.stringEscaper(fmt.Sprintf("%v", val))
				if vstr != "" {
					if _, err := out.Write([]byte(strings.Repeat(opts.ComponentSeparator, compsBefore+1))); err != nil {
						return err
					}
					compsBefore = 0
					if _, err := out.Write([]byte(vstr)); err != nil {
						return err
					}
					continue
				}
			}

			compsBefore++
		}
	}

	_, err = out.Write([]byte(opts.SegmentSeparator))
	if err != nil {
		return err
	}
	return nil
}

func createStringEscaper(opts *Options) (func(input string) string, error) {
	// Escape each character to ensure they are regex-safe
	items := []string{
		regexp.QuoteMeta(opts.SegmentSeparator),
		regexp.QuoteMeta(opts.ComponentSeparator),
		regexp.QuoteMeta(opts.ElementSeparator),
		regexp.QuoteMeta(opts.EscapeCharacter),
	}

	// Join the escaped characters into a regex pattern
	pattern := strings.Join(items, "|")
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(input string) string {
		// Use ReplaceAllStringFunc to replace each match with the prefix + original character
		return re.ReplaceAllStringFunc(input, func(match string) string {
			return opts.EscapeCharacter + match
		})
	}, nil
}

func writeComponent(p *model.Property, data any, out io.Writer, opts *Options) error {
	if p.XEdi.Type != model.EdiTypeComponent {
		return fmt.Errorf("%w: expected 'component', received '%s'", ErrUnexpectedEDIType, p.XEdi.Type)
	}
	if data == nil {
		return nil
	}
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: expected data type to be map[string]any, got %T", ErrUnexpectedInputType, data)
	}
	items := getOrderedEdi(p.Properties)
	var elms []string
	for _, item := range items {
		switch item.Property.XEdi.Type {
		case model.EdiTypeElement:
			if item.Property.Const != nil {
				elms = append(elms, opts.stringEscaper(fmt.Sprintf("%v", item.Property.Const)))
				continue
			}
			if val, ok := m[item.Name]; ok {
				v := opts.stringEscaper(fmt.Sprintf("%v", val))
				if len(v) > 0 {
					elms = append(elms, v)
					continue
				}
			}
			elms = append(elms, "")
		}
	}

	for i := len(elms); i > 0; i-- {
		if elms[i-1] == "" {
			elms = append(elms[:i-1])
		} else {
			break
		}
	}

	if _, err := out.Write([]byte(strings.Join(elms, opts.ElementSeparator))); err != nil {
		return err
	}
	return nil
}
