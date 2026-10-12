package handlers

import (
	"context"
	stdErrors "errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/baking-bad/bcdhub/cmd/api/validations"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/mock"
	mock_contract "github.com/baking-bad/bcdhub/internal/models/mock/contract"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	testViewContract = "KT1B74ywpG3nX2F3ZgpYi4MnsyuoYyQtQTAn"
	testOnChainViews = `[{"prim":"view","args":[{"string":"get_price"},{"prim":"string"},{"prim":"nat"},[{"prim":"CDR"}]]}]`
)

func registerTestValidations(t *testing.T) {
	t.Helper()

	v, ok := binding.Validator.Engine().(*validator.Validate)
	require.True(t, ok)
	require.NoError(t, validations.Register(v, config.APIConfig{Networks: []string{"mainnet"}}))
}

func TestGetErrorCode(t *testing.T) {
	errRecordNotFound := errors.New("record not found")

	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "record not found",
			err:  errRecordNotFound,
			want: http.StatusNotFound,
		}, {
			name: "validation",
			err:  consts.ErrValidation,
			want: http.StatusBadRequest,
		}, {
			name: "wrapped validation",
			err:  errors.Wrap(consts.ErrValidation, "field"),
			want: http.StatusBadRequest,
		}, {
			name: "JSON data is absent",
			err:  consts.ErrJSONDataIsAbsent,
			want: http.StatusBadRequest,
		}, {
			name: "invalid type",
			err:  consts.ErrInvalidType,
			want: http.StatusBadRequest,
		}, {
			name: "no views",
			err:  errNoViews,
			want: http.StatusNotFound,
		}, {
			name: "wrapped no views",
			err:  errors.Wrap(errNoViews, "get_price"),
			want: http.StatusNotFound,
		}, {
			name: "empty implementation",
			err:  errEmptyImplementation,
			want: http.StatusBadRequest,
		}, {
			name: "empty off-chain view",
			err:  errEmptyOffChainView,
			want: http.StatusBadRequest,
		}, {
			name: "timeout",
			err:  context.DeadlineExceeded,
			want: http.StatusGatewayTimeout,
		}, {
			name: "canceled",
			err:  context.Canceled,
			want: statusClientClosedRequest,
		}, {
			name: "wrapped timeout",
			err:  errors.Wrap(context.DeadlineExceeded, "rpc"),
			want: http.StatusGatewayTimeout,
		}, {
			name: "net timeout",
			err:  &net.DNSError{IsTimeout: true},
			want: http.StatusGatewayTimeout,
		}, {
			name: "postgres statement timeout",
			err:  &pgconn.PgError{Code: pgQueryCanceled},
			want: http.StatusGatewayTimeout,
		}, {
			// errorCodes is ordered: the first matched error wins.
			name: "validation joined with no views",
			err:  stdErrors.Join(errNoViews, consts.ErrValidation),
			want: http.StatusBadRequest,
		}, {
			name: "unknown",
			err:  errors.New("unknown"),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mock.NewMockGeneralRepository(ctrl)
			repo.EXPECT().
				IsRecordNotFound(gomock.Any()).
				DoAndReturn(func(err error) bool {
					return errors.Is(err, errRecordNotFound)
				}).
				AnyTimes()

			require.Equal(t, tt.want, getErrorCode(tt.err, repo))
		})
	}
}

func TestExecuteViewRequest_Binding(t *testing.T) {
	registerTestValidations(t)

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name: "on-chain",
			body: `{"data":{"@string_1":"USDTUSD"},"implementation":0,"kind":"on-chain","name":"get_price"}`,
		}, {
			name: "off-chain",
			body: `{"data":{},"kind":"off-chain","name":"get_price","view":{"michelsonStorageView":{"returnType":{"prim":"nat"},"code":[]}}}`,
		}, {
			// Regression test: kind with underscore passed validation and failed with 500.
			name:    "on_chain with underscore",
			body:    `{"data":{"@string_1":"USDTUSD"},"implementation":0,"kind":"on_chain","name":"get_price"}`,
			wantErr: true,
		}, {
			name:    "unknown kind",
			body:    `{"data":{},"implementation":0,"kind":"unknown","name":"get_price"}`,
			wantErr: true,
		}, {
			name:    "empty kind",
			body:    `{"data":{},"implementation":0,"kind":"","name":"get_price"}`,
			wantErr: true,
		}, {
			name:    "without kind",
			body:    `{"data":{},"implementation":0,"name":"get_price"}`,
			wantErr: true,
		}, {
			name:    "on-chain without name",
			body:    `{"data":{},"implementation":0,"kind":"on-chain"}`,
			wantErr: true,
		}, {
			name:    "on-chain without implementation",
			body:    `{"data":{},"kind":"on-chain","name":"get_price"}`,
			wantErr: true,
		}, {
			name:    "off-chain without view",
			body:    `{"data":{},"kind":"off-chain","name":"get_price"}`,
			wantErr: true,
		}, {
			name:    "without data",
			body:    `{"implementation":0,"kind":"on-chain","name":"get_price"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req executeViewRequest
			err := binding.JSON.BindBody([]byte(tt.body), &req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestGetViewForExecute(t *testing.T) {
	tests := []struct {
		name       string
		req        executeViewRequest
		rawViews   string
		wantErr    error
		wantParams string
	}{
		{
			name: "on-chain",
			req: executeViewRequest{
				Kind: OnchainView,
				Name: "get_price",
				Data: map[string]any{"@string_1": "USDTUSD"},
			},
			rawViews:   testOnChainViews,
			wantParams: `{"string":"USDTUSD"}`,
		}, {
			// Regression test: contract without views returned nil view and nil error.
			name: "on-chain: contract without views",
			req: executeViewRequest{
				Kind: OnchainView,
				Name: "get_price",
				Data: map[string]any{},
			},
			rawViews: `[]`,
			wantErr:  errNoViews,
		}, {
			name: "on-chain: unknown view name",
			req: executeViewRequest{
				Kind: OnchainView,
				Name: "unknown",
				Data: map[string]any{},
			},
			rawViews: testOnChainViews,
			wantErr:  errNoViews,
		}, {
			name: "off-chain: nil view",
			req: executeViewRequest{
				Kind: OffchainView,
				Name: "get_price",
				Data: map[string]any{},
			},
			wantErr: errEmptyOffChainView,
		}, {
			name: "off-chain: null implementation",
			req: executeViewRequest{
				Kind: OffchainView,
				Name: "get_price",
				Data: map[string]any{},
				View: &contract.ViewImplementation{
					MichelsonStorageView: contract.Sections{
						Parameter:  []byte("null"),
						ReturnType: []byte("null"),
						Code:       []byte("null"),
					},
				},
			},
			wantErr: errEmptyImplementation,
		}, {
			name: "off-chain: zero value implementation",
			req: executeViewRequest{
				Kind: OffchainView,
				Name: "get_price",
				Data: map[string]any{},
				View: &contract.ViewImplementation{},
			},
			wantErr: errEmptyImplementation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			contracts := mock_contract.NewMockRepository(ctrl)
			if tt.rawViews != "" {
				contracts.EXPECT().
					ScriptPart(gomock.Any(), testViewContract, gomock.Any(), consts.VIEWS).
					Return([]byte(tt.rawViews), nil).
					Times(1)
			}

			view, parameters, err := getViewForExecute(
				context.Background(),
				&config.Context{Contracts: contracts},
				testViewContract,
				tt.req,
			)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, view)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, view)
			require.JSONEq(t, tt.wantParams, string(parameters))
		})
	}

	t.Run("invalid kind", func(t *testing.T) {
		view, _, err := getViewForExecute(
			context.Background(),
			&config.Context{},
			testViewContract,
			executeViewRequest{Kind: "on_chain"},
		)
		require.Error(t, err)
		require.Nil(t, view)
	})
}

func TestExecuteView_ErrorCodes(t *testing.T) {
	registerTestValidations(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		body        string
		rawViews    string
		wantCode    int
		wantMessage string
	}{
		{
			// Regression test: the request from sentry event, it was 500.
			name:     "on_chain with underscore",
			body:     `{"data":{"string":"USDTUSD"},"implementation":0,"kind":"on_chain","name":"get_price"}`,
			wantCode: http.StatusBadRequest,
		}, {
			name:        "contract without views",
			body:        `{"data":{},"implementation":0,"kind":"on-chain","name":"get_price"}`,
			rawViews:    `[]`,
			wantCode:    http.StatusNotFound,
			wantMessage: errNoViews.Error(),
		}, {
			name:        "unknown view name",
			body:        `{"data":{},"implementation":0,"kind":"on-chain","name":"unknown"}`,
			rawViews:    testOnChainViews,
			wantCode:    http.StatusNotFound,
			wantMessage: errNoViews.Error(),
		}, {
			name:     "off-chain without view",
			body:     `{"data":{},"kind":"off-chain","name":"get_price"}`,
			wantCode: http.StatusBadRequest,
		}, {
			name:        "off-chain with null implementation",
			body:        `{"data":{},"kind":"off-chain","name":"get_price","view":{"michelsonStorageView":{"parameter":null,"returnType":null,"code":null}}}`,
			wantCode:    http.StatusBadRequest,
			wantMessage: errEmptyImplementation.Error(),
		}, {
			name:        "off-chain with absent implementation",
			body:        `{"data":{},"kind":"off-chain","name":"get_price","view":{}}`,
			wantCode:    http.StatusBadRequest,
			wantMessage: errEmptyImplementation.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := mock.NewMockGeneralRepository(ctrl)
			storage.EXPECT().IsRecordNotFound(gomock.Any()).Return(false).AnyTimes()

			contracts := mock_contract.NewMockRepository(ctrl)
			if tt.rawViews != "" {
				contracts.EXPECT().
					ScriptPart(gomock.Any(), testViewContract, gomock.Any(), consts.VIEWS).
					Return([]byte(tt.rawViews), nil).
					Times(1)
			}

			cfgCtx := &config.Context{
				Storage:   storage,
				Contracts: contracts,
			}

			router := gin.New()
			router.POST("/v1/contract/:network/:address/views/execute", func(c *gin.Context) {
				c.Set("context", cfgCtx)
			}, ExecuteView())

			request := httptest.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/v1/contract/mainnet/"+testViewContract+"/views/execute",
				strings.NewReader(tt.body),
			)
			request.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			require.Equal(t, tt.wantCode, recorder.Code, recorder.Body.String())

			if tt.wantMessage != "" {
				var response Error
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
				require.Equal(t, tt.wantMessage, response.Message)
			}
		})
	}
}
