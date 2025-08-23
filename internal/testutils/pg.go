package testutils

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	embedded "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("getFreePort listen: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func StartEmbeddedPG(t *testing.T) (*embedded.EmbeddedPostgres, *pgxpool.Pool) {
	t.Helper()
	port := getFreePort(t)

	baseDir := filepath.Join(os.TempDir(), fmt.Sprintf("epg-%d", time.Now().UnixNano()))
	binDir := filepath.Join(baseDir, "bin")
	dataDir := filepath.Join(baseDir, "data")

	cfg := embedded.DefaultConfig().
		Username("postgres").
		Password("secret").
		Database("testdb").
		Version(embedded.V13).
		Port(uint32(port)).
		BinariesPath(binDir).
		DataPath(dataDir)

	epg := embedded.NewDatabase(cfg)
	if err := epg.Start(); err != nil {
		t.Fatalf("embedded start: %v", err)
	}

	uri := fmt.Sprintf("postgres://postgres:secret@127.0.0.1:%d/testdb?sslmode=disable", port)

	var pool *pgxpool.Pool
	deadline := time.Now().Add(8 * time.Second)
	for {
		var err error
		pool, err = pgxpool.New(context.Background(), uri)
		if err == nil {
			if pingErr := pool.Ping(context.Background()); pingErr == nil {
				break
			}
			pool.Close()
		}
		if time.Now().After(deadline) {
			_ = epg.Stop()
			t.Fatalf("cannot connect to embedded postgres: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	t.Cleanup(func() {
		pool.Close()
		_ = epg.Stop()
		os.RemoveAll(baseDir)
	})

	return epg, pool
}
