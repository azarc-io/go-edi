package edi_test

import (
	"encoding/json"
	"github.com/azarc-io/go-edi/pkg/edi"
	"github.com/azarc-io/go-edi/pkg/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		Name     string
		Schema   []byte
		JSON     string
		Expected string
	}{
		{
			Name:     "One Item",
			Schema:   sample1,
			JSON:     `{"foo":{"actionCoded":"0083","applicationRecipient":{"identification":"S007","identificationCodeQualifier":"0044"}}}`,
			Expected: `FOO+0083+S007:0044'`,
		},
		{
			Name:     "Two Items",
			Schema:   sample1,
			JSON:     `{"bar":{"actionCoded":"0083"},"foo":{"actionCoded":"8877","applicationRecipient":{"identification":"T002","identificationCodeQualifier":"6655"}}}`,
			Expected: `BAR+0083'FOO+8877+T002:6655'`,
		},
		{
			Name:     "Array: Object with Object and Array properties",
			Schema:   sample2,
			JSON:     `{"bar":{"actionCoded":"8877"},"foo":[{"actionCoded":"A0001","applicationRecipient":{"identification":"B0001","identificationCodeQualifier":"C0001"}},{"actionCoded":"A0002","applicationRecipient":{"identification":"B0002","identificationCodeQualifier":"C0002"}},{"actionCoded":"A0003","applicationRecipient":{"identification":"B0003","identificationCodeQualifier":"C0003"}}]}`,
			Expected: `BAR+8877'FOO+A0001+B0001:C0001'FOO+A0002+B0002:C0002'FOO+A0003+B0003:C0003'`,
		},
		{
			Name:     "Array: Array with max items and two objects properties",
			Schema:   sample2,
			JSON:     `{"bar":{"actionCoded":"8877"},"foo":[{"actionCoded":"A0001","applicationRecipient":{"identification":"B0001","identificationCodeQualifier":"C0001"}},{"actionCoded":"A0002","applicationRecipient":{"identification":"B0002","identificationCodeQualifier":"C0002"}},{"actionCoded":"A0003","applicationRecipient":{"identification":"B0003","identificationCodeQualifier":"C0003"}}],"foo-next":{"actionCoded":"A0004"}}`,
			Expected: `BAR+8877'FOO+A0001+B0001:C0001'FOO+A0002+B0002:C0002'FOO+A0003+B0003:C0003'FOO+A0004'`,
		},
		{
			Name:     "Complex: Nested Objects",
			Schema:   sample3,
			JSON:     `{"bar":{"actionCoded":"8877"},"foo":[{"actionCoded":"A0001"},{"actionCoded":"A0002"},{"actionCoded":"A0003"}],"foo-next":{"actionCoded":"A0004"},"welcome":{"accountDetails":{"my-bool":true,"my-int":12,"my-number":87.23,"my-string":"my-value"},"clientDetails":{"address":{"line1":"My Line1","line2":"My Line2","postcode":"MyPC"},"name":"Andrew"},"users":[{"emails":[{"email":"t@t1.com"},{"email":"t@t2.com"}],"firstName":"Sam","lastName":"Goodwin"}]}}`,
			Expected: `NAD+Andrew+My Line1:My Line2:MyPC'USR+Sam+Goodwin'EML+t@t1.com'EML+t@t2.com'BAR+8877'FOO+A0001'FOO+A0002'FOO+A0003'FOO+A0004'`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			s, err := schemas.LoadSchema(test.Schema)
			require.NoError(t, err)
			m := make(map[string]any)
			err = json.Unmarshal([]byte(test.JSON), &m)
			require.NoError(t, err)
			data, err := edi.Marshal(s, m)
			require.NoError(t, err)
			assert.Equal(t, test.Expected, string(data))
		})
	}

	t.Run("separator overrides", func(t *testing.T) {
		s, err := schemas.LoadSchema(sample1)
		require.NoError(t, err)
		m := make(map[string]any)
		err = json.Unmarshal([]byte(`{"foo":{"actionCoded":"0083","applicationRecipient":{"identification":"S007","identificationCodeQualifier":"0044"}}}`), &m)
		require.NoError(t, err)
		data, err := edi.Marshal(
			s,
			m,
			edi.WithSegmentSeparator("SS"),
			edi.WithComponentSeparator("CC"),
			edi.WithElementSeparator("EE"),
		)
		require.NoError(t, err)
		assert.Equal(t, "FOOCC0083CCS007EE0044SS", string(data))
	})

	t.Run("escape characters", func(t *testing.T) {
		s, err := schemas.LoadSchema(sample1)
		require.NoError(t, err)
		m := make(map[string]any)
		err = json.Unmarshal([]byte(`{"foo":{"actionCoded":"this+and+that","applicationRecipient":{"identification":"why'other+foo:bar?me","identificationCodeQualifier":"my'seg"}}}`), &m)
		require.NoError(t, err)
		data, err := edi.Marshal(
			s,
			m,
			edi.WithSegmentSeparator("'"),
			edi.WithComponentSeparator("+"),
			edi.WithElementSeparator(":"),
			edi.WithEscapeCharacter("?"),
		)
		require.NoError(t, err)
		assert.Equal(t, "FOO+this?+and?+that+why?'other?+foo?:bar??me:my?'seg'", string(data))
	})

	t.Run("invalid type", func(t *testing.T) {
		s, err := schemas.LoadSchema(sample1)
		require.NoError(t, err)
		data, err := edi.Marshal(s, 100)
		assert.Nil(t, data)
		require.ErrorIs(t, err, edi.ErrUnexpectedInputType)
	})

	t.Run("marshal as struct", func(t *testing.T) {
		type data struct {
			Foo struct {
				ActionCoded          string `json:"actionCoded"`
				ApplicationRecipient struct {
					Identification              string `json:"identification"`
					IdentificationCodeQualifier string `json:"identificationCodeQualifier"`
				} `json:"applicationRecipient"`
			} `json:"foo"`
		}
		var d data
		input := `{"foo":{"actionCoded":"0083","applicationRecipient":{"identification":"S007","identificationCodeQualifier":"0044"}}}`

		s, err := schemas.LoadSchema(sample1)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(input), &d)
		require.NoError(t, err)
		out, err := edi.Marshal(s, &d)
		require.NoError(t, err)
		assert.Equal(t, "FOO+0083+S007:0044'", string(out))
	})
}
