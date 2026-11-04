package handlers

import (
	"net/http"
	"testing"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// A user script with a non-comparable ticket content is a client error: it
// must map to 400, not to 500 with a Sentry event.
func TestForkScript_NonComparableTicketIsBadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockGeneralRepository(ctrl)
	repo.EXPECT().IsRecordNotFound(gomock.Any()).Return(false).AnyTimes()

	for _, storage := range []string{
		`{"prim":"ticket","args":[{"prim":"ticket","args":[{"prim":"nat"}]}]}`,
		`{"prim":"ticket","args":[null]}`,
	} {
		script, err := ast.NewScript([]byte(`[{"prim":"parameter","args":[{"prim":"unit"}]},{"prim":"storage","args":[` + storage + `]},{"prim":"code","args":[[{"prim":"FAILWITH"}]]}]`))
		require.NoError(t, err)

		_, err = script.StorageType()
		require.ErrorIs(t, err, consts.ErrTypeIsNotComparable)
		require.Equal(t, http.StatusBadRequest, getErrorCode(err, repo), storage)
	}
}
