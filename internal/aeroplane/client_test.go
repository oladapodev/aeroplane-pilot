package aeroplane

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListServicesAndDeployments(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "aeroplane.db")
	writeFixtureDB(t, dbPath)

	client, err := NewClient(dbPath)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	services, err := client.ListServices()
	if err != nil {
		t.Fatalf("ListServices() error = %v", err)
	}
	if len(services) != 2 {
		t.Fatalf("len(services) = %d, want 2", len(services))
	}
	if services[0].Name != "api" || services[1].Name != "web" {
		t.Fatalf("unexpected services: %#v", services)
	}

	deployments, err := client.LatestDeployments(1)
	if err != nil {
		t.Fatalf("LatestDeployments() error = %v", err)
	}
	if len(deployments) != 1 {
		t.Fatalf("len(deployments) = %d, want 1", len(deployments))
	}
	if deployments[0].Status != "failed" {
		t.Fatalf("deployments[0].Status = %q, want failed", deployments[0].Status)
	}

	failed, err := client.FailedDeployments(5)
	if err != nil {
		t.Fatalf("FailedDeployments() error = %v", err)
	}
	if len(failed) != 1 || failed[0].ID != "dep-2" {
		t.Fatalf("failed deployments = %#v, want dep-2", failed)
	}

	logs, err := client.DeploymentLogs("dep-2", 10)
	if err != nil {
		t.Fatalf("DeploymentLogs() error = %v", err)
	}
	if len(logs) != 2 || logs[0] != "build started" || logs[1] != "build failed" {
		t.Fatalf("logs = %#v", logs)
	}
}

func writeFixtureDB(t *testing.T, path string) {
	t.Helper()

	sql := `
CREATE TABLE projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  runtime_mode TEXT NOT NULL,
  last_deployed_at TEXT
);
CREATE TABLE deployments (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL,
  status TEXT NOT NULL,
  trigger TEXT NOT NULL,
  started_at TEXT,
  finished_at TEXT,
  created_at TEXT NOT NULL
);
CREATE TABLE deployment_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  deployment_id TEXT NOT NULL,
  line TEXT NOT NULL,
  created_at TEXT NOT NULL
);
INSERT INTO projects (id, name, status, runtime_mode, last_deployed_at) VALUES
  ('svc-1', 'api', 'running', 'web', '2026-06-09T00:00:00Z'),
  ('svc-2', 'web', 'running', 'static', '2026-06-09T01:00:00Z');
INSERT INTO deployments (id, service_id, status, trigger, started_at, finished_at, created_at) VALUES
  ('dep-1', 'svc-1', 'succeeded', 'push', '2026-06-09T00:10:00Z', '2026-06-09T00:12:00Z', '2026-06-09T00:10:00Z'),
  ('dep-2', 'svc-2', 'failed', 'manual', '2026-06-09T01:10:00Z', '2026-06-09T01:11:00Z', '2026-06-09T01:10:00Z');
INSERT INTO deployment_logs (deployment_id, line, created_at) VALUES
  ('dep-2', 'build started', '2026-06-09T01:10:01Z'),
  ('dep-2', 'build failed', '2026-06-09T01:10:02Z');
`
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(path)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.Close()
	if err := execSQL(path, sql); err != nil {
		t.Fatalf("execSQL() error = %v", err)
	}
}
