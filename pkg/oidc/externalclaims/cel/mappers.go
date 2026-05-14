package cel

import (
	"context"

	"github.com/google/cel-go/common/types/traits"
	authenticationcel "github.com/everettraven/padlok/pkg/internal/third_party/kubernetes/apiserver/pkg/authentication/cel"
)

// ExternalSourceCELMapper is a struct that holds the compiled expressions
// used when externally sourcing claims.
type ExternalSourceCELMapper struct {
	URL        authenticationcel.ClaimsMapper
	Conditions authenticationcel.ClaimsMapper
	Sources    ExternalClaimsMapper
}

func NewExternalClaimsMapper(compilationResults ...authenticationcel.CompilationResult) ExternalClaimsMapper {
	return &externalClaimsMapper{
		mapper: authenticationcel.NewMapper(compilationResults),
	}
}

type externalClaimsMapper struct {
	mapper *authenticationcel.Mapper
}

// EvalExternalClaims evaluates the given external claims and returns a list of EvaluationResult.
// This is used for external claim source validation that contains multiple external claims.
func (ecm *externalClaimsMapper) EvalExternalClaims(ctx context.Context, input traits.Mapper) ([]authenticationcel.EvaluationResult, error) {
	return ecm.mapper.Eval(ctx, authenticationcel.NewVarNameActivation(responseVarName, input))
}
