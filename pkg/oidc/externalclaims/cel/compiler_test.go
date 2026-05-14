package cel

import (
	"strings"
	"testing"
)

func TestCompileExternalSourceExpression(t *testing.T) {
	type testcase struct {
		name       string
		expression string
		err        string
	}

	testcases := []testcase{
		{
			name:       "valid expression, uses response variable, optional key presence check, successfully compiles",
			expression: "response.?key.orValue(\"\")",
		},
		{
			name:       "valid expression, uses response variable, successfully compiles",
			expression: "response.key.orValue(\"\")",
		},
		{
			name:       "valid expression, does not use response variable, successfully compiles",
			expression: "'static'",
		},
		{
			name:       "invalid expression, uses claims variable, does not compile",
			expression: "claims.sub",
			err:        "undeclared reference to 'claims'",
		},
		{
			name:       "invalid expression, uses user variable, does not compile",
			expression: "user.groups",
			err:        "undeclared reference to 'user'",
		},
		{
			name:       "invalid expression, uses non-existent variable, does not compile",
			expression: "notreal.key",
			err:        "undeclared reference to 'notreal'",
		},
		{
			name:       "invalid expression, does not compile",
			expression: "!@&^()&*",
			err:        "Syntax error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expressionAccessor := ExternalSourceMappingExpression{
				Claim:      "test",
				Expression: tc.expression,
			}

			compiler := NewCompiler()
			_, err := compiler.CompileExternalSourceExpression(&expressionAccessor)
			switch {
			case err == nil && len(tc.err) > 0:
				t.Fatalf("expected error containing %q but got none", tc.err)
			case err != nil && len(tc.err) == 0:
				t.Fatalf("received an unexpected error: %v", err)
			case err != nil && len(tc.err) > 0 && !strings.Contains(err.Error(), tc.err):
				t.Fatalf("error %v does not contain expected substring %q", err, tc.err)
			}
		})
	}
}
