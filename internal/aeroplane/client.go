package aeroplane

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Service struct {
	ID             string
	Name           string
	Status         string
	RuntimeMode    string
	LastDeployedAt sql.NullString
}

type Deployment struct {
	ID        string
	ServiceID string
	Status    string
	Trigger   string
	StartedAt sql.NullString
	FinishedAt sql.NullString
}

type Client struct {
	db *sql.DB
}

func NewClient(path string) (*Client, error) {
	db, err := openDB(path)
	if err != nil {
		return nil, err
	}
	return &Client{db: db}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) ListServices() ([]Service, error) {
	rows, err := c.db.Query(`SELECT id, name, status, runtime_mode, last_deployed_at FROM projects ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var svc Service
		if err := rows.Scan(&svc.ID, &svc.Name, &svc.Status, &svc.RuntimeMode, &svc.LastDeployedAt); err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	return services, rows.Err()
}

func (c *Client) LatestDeployments(limit int) ([]Deployment, error) {
	rows, err := c.db.Query(`SELECT id, service_id, status, trigger, started_at, finished_at FROM deployments ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var dep Deployment
		if err := rows.Scan(&dep.ID, &dep.ServiceID, &dep.Status, &dep.Trigger, &dep.StartedAt, &dep.FinishedAt); err != nil {
			return nil, err
		}
		deployments = append(deployments, dep)
	}
	return deployments, rows.Err()
}

func (c *Client) FailedDeployments(limit int) ([]Deployment, error) {
	rows, err := c.db.Query(`SELECT id, service_id, status, trigger, started_at, finished_at FROM deployments WHERE status = 'failed' ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var dep Deployment
		if err := rows.Scan(&dep.ID, &dep.ServiceID, &dep.Status, &dep.Trigger, &dep.StartedAt, &dep.FinishedAt); err != nil {
			return nil, err
		}
		deployments = append(deployments, dep)
	}
	return deployments, rows.Err()
}

func (c *Client) DeploymentLogs(deploymentID string, maxLines int) ([]string, error) {
	rows, err := c.db.Query(`SELECT line FROM deployment_logs WHERE deployment_id = ? ORDER BY id DESC LIMIT ?`, deploymentID, maxLines)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines, rows.Err()
}
