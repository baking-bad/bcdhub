package core

import (
	"testing"
	"time"

	pg "github.com/go-pg/pg/v10"
	"github.com/stretchr/testify/assert"
)

func Test_parseConncetionString(t *testing.T) {
	tests := []struct {
		name       string
		connection string
		want       *pg.Options
		wantErr    bool
	}{
		{
			name:       "test 1",
			connection: "host=127.0.0.1 port=5432 user=user dbname=indexer password=password sslmode=disable",
			want: &pg.Options{
				Addr:               "127.0.0.1:5432",
				User:               "user",
				Password:           "password",
				Database:           "indexer",
				IdleTimeout:        time.Second * 15,
				IdleCheckFrequency: time.Second * 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConnectionString(tt.connection, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("parseConncetionString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.Addr, got.Addr)
			assert.Equal(t, tt.want.User, got.User)
			assert.Equal(t, tt.want.Password, got.Password)
			assert.Equal(t, tt.want.Database, got.Database)
			assert.Equal(t, tt.want.IdleTimeout, got.IdleTimeout)
			assert.Equal(t, tt.want.IdleCheckFrequency, got.IdleCheckFrequency)
		})
	}
}
