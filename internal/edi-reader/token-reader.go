package edi_reader

import (
	"bytes"
)

type ediTokenReader struct {
	data      []byte
	separator string
	escape    string
	cursor    int
}

func (tr *ediTokenReader) ReadNext() ([]byte, bool) {
	out, end := tr.PeekNext()
	tr.cursor = end + 1
	return out, tr.cursor >= len(tr.data)
}

func (tr *ediTokenReader) PeekNext() ([]byte, int) {
	// Find the next Segment delimiter
	var end int
	var start = tr.cursor

	if start >= len(tr.data) {
		// already completed
		return nil, -1
	}

	for {
		end = bytes.IndexAny(tr.data[start:], tr.separator)
		if end == -1 {
			// end of dataset
			end = len(tr.data)
			break
		}

		nextCursor := end + start
		if nextCursor > 0 && string(tr.data[nextCursor-1]) == tr.escape {
			// this is an escape char, continue
			// remove escape character before moving on
			tr.data = removeByteAtPosition(tr.data, nextCursor-1)
			start = nextCursor
		} else {
			// found next separator
			end = nextCursor
			break
		}
	}

	return tr.data[tr.cursor:end], end
}

func removeByteAtPosition(data []byte, k int) []byte {
	if k < 0 || k >= len(data) {
		// If k is out of bounds, return the original slice
		return data
	}
	// Remove the byte at position K by concatenating the slices before and after K
	return append(data[:k], data[k+1:]...)
}

func NewEdiTokenReader(data []byte, separator, escape string) *ediTokenReader {
	return &ediTokenReader{
		data:      data,
		separator: separator,
		escape:    escape,
	}
}
