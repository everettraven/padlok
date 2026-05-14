package cel

import (
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestEvalExternalClaim(t *testing.T) {
	type testcase struct {
		name     string
		accessor *ExternalSourceMappingExpression
		input    map[string]any
		resultFn func(*testing.T, ref.Val)
		err      string
	}

	testcases := []testcase{
		{
			name: "simple response input, evaluates to string type",
			accessor: &ExternalSourceMappingExpression{
				Claim:      "test",
				Expression: "response.testkey",
			},
			input: map[string]any{
				"testkey": "testvalue",
			},
			resultFn: func(t *testing.T, v ref.Val) {
				if v == nil || v.Type() == nil {
					t.Fatal("expected a result but got none")
				}

				if v.Type() != cel.StringType {
					t.Fatalf("expected a string typed result but got %v", v.Type())
				}

				if v.Value().(string) != "testvalue" {
					t.Fatalf("expected result value to be %q but got %q", "testvalue", v.Value())
				}
			},
		},
		{
			name: "simple response input, evaluates to list type",
			accessor: &ExternalSourceMappingExpression{
				Claim:      "test",
				Expression: "response.testkey",
			},
			input: map[string]any{
				"testkey": []string{"testvalue1", "testvalue2"},
			},
			resultFn: func(t *testing.T, v ref.Val) {
				if v == nil || v.Type() == nil {
					t.Fatal("expected a result but got none")
				}

				if v.Type().TypeName() != cel.ListType(cel.DynType).TypeName() {
					t.Fatalf("expected a list typed result but got %v", v.Type().TypeName())
				}

				actual := sets.New(v.Value().([]string)...)
				if !actual.HasAll("testvalue1", "testvalue2") {
					t.Fatal("expected result value to be a list containing entries 'testvalue1' and 'testvalue2' but it did not")
				}
			},
		},
		{
			name: "simple response input, expression attempts to gets non-existent key, fails to evaluate",
			accessor: &ExternalSourceMappingExpression{
				Claim:      "test",
				Expression: "response.nokey",
			},
			input: map[string]any{
				"testkey": "testvalue",
			},
			err: "no such key: nokey",
		},
		{
			name: "simple response input, expression attempts to optionally get non-existent key, evaluates successfully",
			accessor: &ExternalSourceMappingExpression{
				Claim:      "test",
				Expression: "has(response.nokey) ? response.nokey : 'unknown'",
			},
			input: map[string]any{
				"testkey": "testvalue",
			},
			resultFn: func(t *testing.T, v ref.Val) {
				if v == nil || v.Type() == nil {
					t.Fatal("expected a result but got none")
				}

				if v.Type() != cel.StringType {
					t.Fatalf("expected a string typed result but got %v", v.Type())
				}

				if v.Value().(string) != "unknown" {
					t.Fatalf("expected result value to be %q but got %q", "unknown", v.Value())
				}
			},
		},
		{
			name: "nested response input, expression gets nested key, evaluates successfully",
			accessor: &ExternalSourceMappingExpression{
				Claim:      "test",
				Expression: "response.keys.nested",
			},
			input: map[string]any{
				"keys": map[string]any{
					"nested": "testvalue",
				},
			},
			resultFn: func(t *testing.T, v ref.Val) {
				if v == nil || v.Type() == nil {
					t.Fatal("expected a result but got none")
				}

				if v.Type() != cel.StringType {
					t.Fatalf("expected a string typed result but got %v", v.Type())
				}

				if v.Value().(string) != "testvalue" {
					t.Fatalf("expected result value to be %q but got %q", "testvalue", v.Value())
				}
			},
		},
		{
			name: "nested response input, expression gets top-level key, evaluates successfully",
			accessor: &ExternalSourceMappingExpression{
				Claim:      "test",
				Expression: "response.keys",
			},
			input: map[string]any{
				"keys": map[string]any{
					"nested": "testvalue",
				},
			},
			resultFn: func(t *testing.T, v ref.Val) {
				if v == nil || v.Type() == nil {
					t.Fatal("expected a result but got none")
				}

				if v.Type().TypeName() != cel.MapType(cel.DynType, cel.DynType).TypeName() {
					t.Fatalf("expected a map typed result but got %v", v.Type().TypeName())
				}
			},
		},
		{
			name: "complex response input, complex expression, evaluates to string result successfully",
			accessor: &ExternalSourceMappingExpression{
				Claim:      "test",
				Expression: "response.?groups.orValue([]).map(x, x.name).join(',')",
			},
			input: map[string]any{
				"groups": []map[string]any{
					{
						"name": "humans",
						"id":   123456,
					},
					{
						"name": "developers",
						"id":   7891011,
					},
				},
			},
			resultFn: func(t *testing.T, v ref.Val) {
				if v == nil || v.Type() == nil {
					t.Fatal("expected a result but got none")
				}

				if v.Type() != cel.StringType {
					t.Fatalf("expected a string typed result but got %v", v.Type())
				}

				if v.Value().(string) != "humans,developers" {
					t.Fatalf("expected result value to be %q but got %q", "humans,developers", v.Value())
				}

			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			compiler := NewCompiler()

			out, err := compiler.CompileExternalSourceExpression(tc.accessor)
			if err != nil {
				t.Fatalf("expression %q failed to compile: %v", tc.accessor.Expression, err)
			}

			mapper := NewExternalClaimsMapper(out)
			input := types.NewStringInterfaceMap(types.DefaultTypeAdapter, tc.input)

			result, err := mapper.EvalExternalClaims(t.Context(), input)
			switch {
			case err == nil && len(tc.err) > 0:
				t.Fatalf("expected error containing %q but got none", tc.err)
			case err != nil && len(tc.err) == 0:
				t.Fatalf("received an unexpected error: %v", err)
			case err != nil && len(tc.err) > 0 && !strings.Contains(err.Error(), tc.err):
				t.Fatalf("error %v does not contain expected substring %q", err, tc.err)
			}

			if tc.resultFn != nil {
				tc.resultFn(t, result[0].EvalResult)
			}
		})
	}
}
