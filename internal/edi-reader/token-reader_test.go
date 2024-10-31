package edi_reader_test

import (
	"testing"

	edi_reader "github.com/azarc-io/go-edi/internal/edi-reader"
	"github.com/stretchr/testify/assert"
)

func TestEdiTokenReader(t *testing.T) {
	t.Run("read to next item", func(t *testing.T) {
		data := []byte(`this is+my next?+token here+please`)
		tr := edi_reader.NewEdiTokenReader(data, '+', '?')
		o, _ := tr.ReadNext()
		assert.Equal(t, "this is", string(o))
		o, _ = tr.ReadNext()
		assert.Equal(t, "my next+token here", string(o))
		o, _ = tr.ReadNext()
		assert.Equal(t, "please", string(o))
	})

	t.Run("read none found", func(t *testing.T) {
		data := []byte(`this is+my next?+token here+please`)
		tr := edi_reader.NewEdiTokenReader(data, '&', '?')
		o, finished := tr.ReadNext()
		assert.Equal(t, "this is+my next?+token here+please", string(o))
		assert.True(t, finished)
	})

	t.Run("read at beginning", func(t *testing.T) {
		data := []byte(`+this ?+ that`)
		tr := edi_reader.NewEdiTokenReader(data, '+', '?')
		o, finished := tr.ReadNext()
		assert.Equal(t, "", string(o))
		assert.False(t, finished)
		o, finished = tr.ReadNext()
		assert.Equal(t, "this + that", string(o))
		assert.True(t, finished)
	})

	t.Run("read at end", func(t *testing.T) {
		data := []byte(`this is+`)
		tr := edi_reader.NewEdiTokenReader(data, '+', '?')
		o, finished := tr.ReadNext()
		assert.Equal(t, "this is", string(o))
		assert.True(t, finished)
	})

	t.Run("peek next does not advance", func(t *testing.T) {
		data := []byte(`this is + just like that`)
		tr := edi_reader.NewEdiTokenReader(data, '+', '?')
		o, _ := tr.PeekNext()
		assert.Equal(t, "this is ", string(o))
		o, _ = tr.PeekNext()
		assert.Equal(t, "this is ", string(o))

		_, _ = tr.ReadNext()
		o, _ = tr.PeekNext()
		assert.Equal(t, " just like that", string(o))
		o, _ = tr.PeekNext()
		assert.Equal(t, " just like that", string(o))
	})

	t.Run("peek next already complete", func(t *testing.T) {
		data := []byte(`this is`)
		tr := edi_reader.NewEdiTokenReader(data, '+', '?')
		o, _ := tr.PeekNext()
		assert.Equal(t, "this is", string(o))
		_, _ = tr.ReadNext()
		o, i := tr.PeekNext()
		assert.Equal(t, "", string(o))
		assert.Equal(t, -1, i)
	})
}
