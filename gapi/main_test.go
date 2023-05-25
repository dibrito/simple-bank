package gapi

import (
	"testing"
	"time"

	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/util"
	"github.com/stretchr/testify/require"

	"github.com/dibrito/simple-bank/worker"
)

// NewServer creates a new gRPC server.
func newTestServer(t *testing.T, store db.Store, td worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey: util.RandomString(32),
		AccessDuration:    time.Minute,
	}
	server, err := NewServer(config, store, td)
	require.NoError(t, err)
	return server
}
