package faketagencoder

import (
	"reflect"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestTag(t *testing.T) {
	type testCase struct {
		input    reflect.StructTag
		tag      string
		option   string
		expected string
	}

	for _, tc := range []testCase{
		{
			input:    `json:"foo"`,
			tag:      `json`,
			option:   `omitempty`,
			expected: `json:"foo,omitempty"`,
		},
		{
			input:    `json:"'\\xde\\xad\\xbe\\xef'"`,
			tag:      `json`,
			option:   `omitzero`,
			expected: `json:"'\\xde\\xad\\xbe\\xef',omitzero"`,
		},
		{
			input:    `json:",omitzero"`,
			tag:      `json`,
			option:   `omitempty`,
			expected: `json:",omitzero,omitempty"`,
		},
		{
			input:    `json:",omitzero"`,
			tag:      `json`,
			option:   `omitzero`,
			expected: `json:",omitzero"`,
		},
		{
			input:    `json:",omitempty"`,
			tag:      `json`,
			option:   `format:booboo`,
			expected: `json:",omitempty,format:booboo"`,
		},
		{
			input:    `json:",format:fizzbuzz"`,
			tag:      `json`,
			option:   `format:booboo`,
			expected: `json:",format:fizzbuzz"`,
		},
		{
			input:    `json:",format:fizzbuzz"`,
			tag:      `json`,
			option:   `omitempty`,
			expected: `json:",format:fizzbuzz,omitempty"`,
		},
		{
			input:    `json:"foo"`,
			tag:      `bar`,
			option:   `baz`,
			expected: `json:"foo" bar:"baz"`,
		},
		{
			input:    `json:"foo"`,
			tag:      `bar`,
			option:   `,baz`,
			expected: `json:"foo" bar:",baz"`,
		},
		{
			input:    `json:"foo" bar:",foo"`,
			tag:      `bar`,
			option:   `baz`,
			expected: `json:"foo" bar:",foo,baz"`,
		},
	} {
		added, err := AddTagOption(tc.input, tc.tag, tc.option)
		assert.NilError(t, err)
		assert.Assert(t, cmp.Equal(reflect.StructTag(tc.expected), added))
	}

}
