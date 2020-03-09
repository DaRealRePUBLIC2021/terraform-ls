package filesystem

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
)

func TestHclBlockAtPos(t *testing.T) {
	testCases := []struct {
		name string

		content string
		pos     hcl.Pos

		expectedErr   error
		expectedBlock *hcl.Block
	}{
		{
			"valid config and position",
			`provider "aws" {

}
`,
			hcl.Pos{
				Line:   2,
				Column: 1,
				Byte:   17,
			},
			nil,
			&hcl.Block{
				Type:   "provider",
				Labels: []string{"aws"},
			},
		},
		{
			"empty config and valid position",
			``,
			hcl.Pos{
				Line:   1,
				Column: 1,
				Byte:   0,
			},
			&NoBlockFoundErr{AtPos: hcl.Pos{Line: 1, Column: 1, Byte: 0}},
			nil,
		},
		{
			"empty config and out-of-range negative position",
			``,
			hcl.Pos{
				Line:   -42,
				Column: -3,
				Byte:   -46,
			},
			&InvalidHclPosErr{
				Pos:     hcl.Pos{Line: -42, Column: -3, Byte: -46},
				InRange: hcl.Range{Filename: "/test.tf", Start: hcl.InitialPos, End: hcl.InitialPos},
			},
			nil,
		},
		{
			"empty config and out-of-range positive position",
			``,
			hcl.Pos{
				Line:   42,
				Column: 3,
				Byte:   46,
			},
			&InvalidHclPosErr{
				Pos:     hcl.Pos{Line: 42, Column: 3, Byte: 46},
				InRange: hcl.Range{Filename: "/test.tf", Start: hcl.InitialPos, End: hcl.InitialPos},
			},
			nil,
		},
		{
			"valid config and out-of-range positive position",
			`provider "aws" {

}
`,
			hcl.Pos{
				Line:   42,
				Column: 3,
				Byte:   46,
			},
			&InvalidHclPosErr{
				Pos: hcl.Pos{Line: 42, Column: 3, Byte: 46},
				InRange: hcl.Range{
					Filename: "/test.tf",
					Start:    hcl.InitialPos,
					End:      hcl.Pos{Column: 1, Line: 4, Byte: 20},
				},
			},
			nil,
		},
	}

	opts := cmpopts.IgnoreFields(hcl.Block{},
		"Body", "DefRange", "TypeRange", "LabelRanges")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i+1, tc.name), func(t *testing.T) {
			uri := "/test.tf"
			f := NewFile(uri, []byte(tc.content))

			block, err := f.HclBlockAtPos(tc.pos)
			if err != nil {
				if tc.expectedErr == nil {
					t.Fatal(err)
				}
				if err.Error() != tc.expectedErr.Error() {
					t.Fatalf("Unexpected error.\nexpected: %#v\ngiven:    %#v\n",
						tc.expectedErr, err)
				}
				return
			}
			if tc.expectedErr != nil {
				t.Fatalf("Expected error: %s", tc.expectedErr)
			}

			if diff := cmp.Diff(block, tc.expectedBlock, opts); diff != "" {
				t.Fatalf("Unexpected block difference: %s", diff)
			}

		})
	}
}
