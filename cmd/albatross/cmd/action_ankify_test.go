package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixLatex(t *testing.T) {
	tcs := []struct {
		in, out string
	}{
		{
			`Q: What is $(3 + 4i)(3 - 4i)$??"`,
			`Q: What is [$](3 + 4i)(3 - 4i)[/$]??"`,
		},
		{
			`What is $z - z^{\ast}$ for $z=(a+bi)$??"`,
			`What is [$]z - z^{\ast}[/$] for [$]z=(a+bi)[/$]??"`,
		},
		{
			`What is the notation for the complex conjugate of $z$??`,
			`What is the notation for the complex conjugate of [$]z[/$]??`,
		},
		{
			`$2a$.`,
			`[$]2a[/$].`,
		},
		{
			`$$\frac{a + bi}{a - bi}}$$`,
			`[$$]\frac{a + bi}{a - bi}}[/$$]`,
		},
		{
			`$$3$$`,
			`[$$]3[/$$]`,
		},
		{
			`$$25$$`,
			`[$$]25[/$$]`,
		},
	}

	for _, tc := range tcs {
		got := fixFlashcardLatex([]string{tc.in}, "[$]", "[/$]", "[$$]", "[/$$]")[0]
		assert.Equal(t, tc.out, got, "expected output latex to be correct")
	}
}
