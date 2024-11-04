package edi_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/azarc-io/go-edi/pkg/edi"
	"github.com/azarc-io/go-edi/pkg/schemas"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed fixtures/sample-1-simple-object.schema.json
var sample1 []byte

//go:embed fixtures/sample-2-simple-array.schema.json
var sample2 []byte

//go:embed fixtures/sample-3-complex.schema.json
var sample3 []byte

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		Name     string
		Schema   []byte
		Edi      string
		Expected string
	}{
		{
			Name:     "One Item",
			Schema:   sample1,
			Edi:      `FOO+0083+S007:0044:0007'`,
			Expected: `{"foo":{"actionCoded":"0083","applicationRecipient":{"identification":"S007","identificationCodeQualifier":"0044"}}}`,
		},
		{
			Name:     "Two Items",
			Schema:   sample1,
			Edi:      `BAR+0083+S007:0044:0007'FOO+8877+T002:6655:3322'`,
			Expected: `{"bar":{"actionCoded":"0083"},"foo":{"actionCoded":"8877","applicationRecipient":{"identification":"T002","identificationCodeQualifier":"6655"}}}`,
		},
		{
			Name:     "Array: Object with Object and Array properties",
			Schema:   sample2,
			Edi:      `BAR+8877+T002:6655:3322'FOO+A0001+B0001:C0001:D0001'FOO+A0002+B0002:C0002:D0002'FOO+A0003+B0003:C0003:D0003'`,
			Expected: `{"bar":{"actionCoded":"8877"},"foo":[{"actionCoded":"A0001","applicationRecipient":{"identification":"B0001","identificationCodeQualifier":"C0001"}},{"actionCoded":"A0002","applicationRecipient":{"identification":"B0002","identificationCodeQualifier":"C0002"}},{"actionCoded":"A0003","applicationRecipient":{"identification":"B0003","identificationCodeQualifier":"C0003"}}]}`,
		},
		{
			Name:     "Array: Array with max items and two objects properties",
			Schema:   sample2,
			Edi:      `BAR+8877+T002:6655:3322'FOO+A0001+B0001:C0001:D0001'FOO+A0002+B0002:C0002:D0002'FOO+A0003+B0003:C0003:D0003'FOO+A0004+B0004:C0004:D0004'`,
			Expected: `{"bar":{"actionCoded":"8877"},"foo":[{"actionCoded":"A0001","applicationRecipient":{"identification":"B0001","identificationCodeQualifier":"C0001"}},{"actionCoded":"A0002","applicationRecipient":{"identification":"B0002","identificationCodeQualifier":"C0002"}},{"actionCoded":"A0003","applicationRecipient":{"identification":"B0003","identificationCodeQualifier":"C0003"}}],"foo-next":{"actionCoded":"A0004"}}`,
		},
		{
			Name:     "Complex: Nested Objects",
			Schema:   sample3,
			Edi:      `NAD+Andrew+My Line1:My Line2:MyPC+'USR+Sam+Goodwin'EML+t@t1.com'EML+t@t2.com'BAR+8877+T002:6655:3322'FOO+A0001+B0001:C0001:D0001'FOO+A0002+B0002:C0002:D0002'FOO+A0003+B0003:C0003:D0003'FOO+A0004+B0004:C0004:D0004'`,
			Expected: `{"bar":{"actionCoded":"8877"},"foo":[{"actionCoded":"A0001"},{"actionCoded":"A0002"},{"actionCoded":"A0003"}],"foo-next":{"actionCoded":"A0004"},"welcome":{"accountDetails":{"my-bool":true,"my-int":12,"my-number":87.23,"my-string":"my-value"},"clientDetails":{"address":{"line1":"My Line1","line2":"My Line2","postcode":"MyPC"},"name":"Andrew"},"users":[{"emails":[{"email":"t@t1.com"},{"email":"t@t2.com"}],"firstName":"Sam","lastName":"Goodwin"}]}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			s, err := schemas.LoadSchema(test.Schema)
			require.NoError(t, err)
			m := make(map[string]any)
			require.NoError(t, err)
			err = edi.Unmarshal(s, []byte(test.Edi), &m)
			require.NoError(t, err)
			l, _ := json.MarshalIndent(m, "", "   ")
			if !assert.JSONEq(t, test.Expected, string(l)) {
				fmt.Println(string(l))
			}
		})
	}
}
