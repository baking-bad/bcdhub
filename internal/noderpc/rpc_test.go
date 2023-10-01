package noderpc

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeRPC_checkStatusCode(t *testing.T) {
	tests := []struct {
		name            string
		r               io.Reader
		statusCode      int
		checkStatusCode bool
		wantErr         bool
		errString       string
	}{
		{
			name:            "404 error",
			r:               strings.NewReader("404 page not found"),
			statusCode:      400,
			checkStatusCode: true,
			wantErr:         true,
			errString:       "404 page not found",
		}, {
			name:            "200",
			r:               strings.NewReader(""),
			statusCode:      200,
			checkStatusCode: true,
			wantErr:         false,
		}, {
			name:            "501",
			r:               strings.NewReader("node unavailiable"),
			statusCode:      501,
			checkStatusCode: true,
			wantErr:         true,
			errString:       "is unavailiable: 501",
		}, {
			name:            "no checking",
			r:               strings.NewReader("node unavailiable"),
			statusCode:      400,
			checkStatusCode: false,
			wantErr:         false,
			errString:       "",
		}, {
			name:            "501 with no checking",
			r:               strings.NewReader("node unavailiable"),
			statusCode:      501,
			checkStatusCode: false,
			wantErr:         true,
			errString:       "is unavailiable: 501",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpc := new(NodeRPC)
			err := rpc.checkStatusCode(tt.r, tt.statusCode, tt.checkStatusCode)
			require.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				require.ErrorContains(t, err, tt.errString)
			}
		})
	}
}
