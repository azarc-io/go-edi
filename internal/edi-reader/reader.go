package edi_reader

import (
	"bytes"
	"fmt"
)

const (
	SeparatorSegmentDefault   = '\''
	SeparatorComponentDefault = '+'
	SeparatorElementDefault   = ':'
	EscapeCharacterDefault    = '?'
)

type EDIReader struct {
	reader             *ediTokenReader
	segmentSeparator   rune
	componentSeparator rune
	elementSeparator   rune
	escapeCharacter    rune
}

type option func(*EDIReader)

func WithSegmentSeparator(s rune) option {
	return func(reader *EDIReader) {
		reader.segmentSeparator = s
	}
}
func WithComponentSeparator(s rune) option {
	return func(reader *EDIReader) {
		reader.componentSeparator = s
	}
}
func WithElementSeparator(s rune) option {
	return func(reader *EDIReader) {
		reader.elementSeparator = s
	}
}
func WithEscapeChar(s rune) option {
	return func(reader *EDIReader) {
		reader.escapeCharacter = s
	}
}

// NewEDIReader initializes the EDIReader with EDI input data
func NewEDIReader(data []byte, opts ...option) *EDIReader {
	er := &EDIReader{
		segmentSeparator:   SeparatorSegmentDefault,
		componentSeparator: SeparatorComponentDefault,
		elementSeparator:   SeparatorElementDefault,
		escapeCharacter:    EscapeCharacterDefault,
	}

	for _, opt := range opts {
		opt(er)
	}

	er.reader = NewEdiTokenReader(data, er.segmentSeparator, er.escapeCharacter)
	return er
}

// ReadSegment reads the next Segment from the EDI data
func (r *EDIReader) ReadSegment() (*Segment, error) {
	data, finished := r.reader.ReadNext()
	if len(data) == 0 && finished {
		return nil, nil
	}

	// Split Segment data into parts and create Segment
	parts := bytes.SplitN(data, []byte{byte(r.componentSeparator)}, 2)
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid Segment format")
	}

	return newSegment(string(bytes.TrimSpace(parts[0])), r, parts[1]), nil
}

// PeekSegment peek the next Segment from the EDI data
func (r *EDIReader) PeekSegment() (string, error) {
	data, finished := r.reader.PeekNext()
	if len(data) == 0 && finished == -1 {
		return "", nil
	}

	// Split Segment data into parts and create Segment
	parts := bytes.SplitN(data, []byte{byte(r.componentSeparator)}, 2)
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid Segment format")
	}

	return string(bytes.TrimSpace(parts[0])), nil
}

// Segment represents a single EDI Segment and holds its components
type Segment struct {
	tag             string
	r               *EDIReader
	componentReader *ediTokenReader
}

// newSegment initializes a Segment with the given tag and components
func newSegment(tag string, r *EDIReader, componentsData []byte) *Segment {
	return &Segment{
		tag:             tag,
		r:               r,
		componentReader: NewEdiTokenReader(componentsData, r.componentSeparator, r.escapeCharacter),
	}
}

// Tag returns the Segment's tag
func (s *Segment) Tag() string {
	return s.tag
}

// ReadComponents returns list of components in the Segment
func (s *Segment) ReadComponents() []*Component {
	var comps []*Component
	for {
		cd, finished := s.componentReader.ReadNext()
		comp := Component{index: len(comps)}
		elReader := NewEdiTokenReader(cd, s.r.elementSeparator, s.r.escapeCharacter)
		for {
			el, finished := elReader.ReadNext()
			comp.elements = append(comp.elements, string(el))
			if finished {
				break
			}
		}

		comps = append(comps, &comp)

		if finished {
			return comps
		}
	}
}

// Component represents a component in an EDI Segment and holds its elements
type Component struct {
	elements []string
	index    int
}

// Elements returns the elements in the component
func (c *Component) Elements() []string {
	return c.elements
}

// Index returns the component index
func (c *Component) Index() int {
	return c.index
}
