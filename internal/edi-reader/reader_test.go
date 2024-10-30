package edi_reader_test

import (
	"testing"

	edi_reader "github.com/azarc-io/go-edi/internal/edi-reader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdiReader(t *testing.T) {
	t.Run("Read", func(t *testing.T) {
		ediData := []byte("UCF+0083:S007+0044:0007'NAD+JOHN+DOE++:2+FOO:BAR'")
		ediReader := edi_reader.NewEDIReader(ediData)

		segment1, err := ediReader.ReadSegment()
		require.NoError(t, err)
		assert.Equal(t, "UCF", segment1.Tag())
		comps := segment1.ReadComponents()
		require.Len(t, comps, 2)
		assert.Equal(t, []string{"0083", "S007"}, comps[0].Elements())
		assert.Equal(t, []string{"0044", "0007"}, comps[1].Elements())

		segment2, err := ediReader.ReadSegment()
		require.NoError(t, err)
		assert.Equal(t, "NAD", segment2.Tag())
		comps = segment2.ReadComponents()
		require.Len(t, comps, 5)
		assert.Equal(t, []string{"JOHN"}, comps[0].Elements())
		assert.Equal(t, []string{"DOE"}, comps[1].Elements())
		assert.Equal(t, []string{""}, comps[2].Elements())
		assert.Equal(t, []string{"", "2"}, comps[3].Elements())
		assert.Equal(t, []string{"FOO", "BAR"}, comps[4].Elements())
	})

	t.Run("Peek", func(t *testing.T) {
		ediData := []byte("UCF+0083:S007+0044:0007'NAD+JOHN+DOE++:2+FOO:BAR'")
		ediReader := edi_reader.NewEDIReader(ediData)

		segment1, err := ediReader.PeekSegment()
		require.NoError(t, err)
		assert.Equal(t, "UCF", segment1)
		segment1, err = ediReader.PeekSegment()
		require.NoError(t, err)
		assert.Equal(t, "UCF", segment1)
	})

	t.Run("With Opts", func(t *testing.T) {
		ediData := []byte("someCgoCthereEwithXEstuffS")
		ediReader := edi_reader.NewEDIReader(ediData,
			edi_reader.WithSegmentSeparator('S'),
			edi_reader.WithComponentSeparator('C'),
			edi_reader.WithElementSeparator('E'),
			edi_reader.WithEscapeChar('X'),
		)

		segment1, err := ediReader.ReadSegment()
		require.NoError(t, err)
		assert.Equal(t, "some", segment1.Tag())
		comps := segment1.ReadComponents()
		require.Len(t, comps, 2)
		assert.Equal(t, []string{"go"}, comps[0].Elements())
		assert.Equal(t, []string{"there", "withEstuff"}, comps[1].Elements())
	})
}
