package core

import (
	"fmt"
	"strings"

	"github.com/oladapodev/aeroplane-pilot/internal/aeroplane"
	"github.com/oladapodev/aeroplane-pilot/internal/llm"
)

type AeroplaneClient interface {
	ListServices() ([]aeroplane.Service, error)
	LatestDeployments(limit int) ([]aeroplane.Deployment, error)
	FailedDeployments(limit int) ([]aeroplane.Deployment, error)
	DeploymentLogs(deploymentID string, maxLines int) ([]string, error)
}

type Router struct {
	client AeroplaneClient
	llm    llm.Generator
}

type Option func(*Router)

func WithLLM(generator llm.Generator) Option {
	return func(r *Router) {
		r.llm = generator
	}
}

func NewRouter(client AeroplaneClient, opts ...Option) *Router {
	r := &Router{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Router) Route(message string) string {
	msg := strings.TrimSpace(message)
	switch msg {
	case "/status":
		return r.status()
	case "/health":
		return r.health()
	case "/recent":
		return r.recent()
	case "/failures":
		return r.failures()
	case "/help", "/commands":
		return r.help()
	default:
		return r.freeText(msg)
	}
}

func (r *Router) status() string {
	services, err := r.client.ListServices()
	if err != nil {
		return fmt.Sprintf("status error: %v", err)
	}
	running, stopped := 0, 0
	for _, s := range services {
		if strings.EqualFold(s.Status, "running") {
			running++
		} else {
			stopped++
		}
	}
	return fmt.Sprintf("status: %d services, %d running, %d stopped", len(services), running, stopped)
}

func (r *Router) health() string {
	services, err := r.client.ListServices()
	if err != nil {
		return fmt.Sprintf("health error: %v", err)
	}
	failed, err := r.client.FailedDeployments(10)
	if err != nil {
		return fmt.Sprintf("health error: %v", err)
	}
	healthy, unhealthy := 0, 0
	for _, s := range services {
		if strings.EqualFold(s.Status, "running") {
			healthy++
		} else {
			unhealthy++
		}
	}
	return fmt.Sprintf("health: %d services, %d healthy, %d unhealthy, %d failed deployments", len(services), healthy, unhealthy, len(failed))
}

func (r *Router) recent() string {
	deps, err := r.client.LatestDeployments(5)
	if err != nil {
		return fmt.Sprintf("recent error: %v", err)
	}
	if len(deps) == 0 {
		return "recent: no deployments found"
	}
	var lines []string
	for _, d := range deps {
		status := d.Status
		if d.FinishedAt.Valid {
			status = fmt.Sprintf("%s (%s)", d.Status, d.FinishedAt.String)
		}
		lines = append(lines, fmt.Sprintf("- %s: %s (%s)", d.ID, d.ServiceID, status))
	}
	return "recent:\n" + strings.Join(lines, "\n")
}

func (r *Router) failures() string {
	deps, err := r.client.FailedDeployments(10)
	if err != nil {
		return fmt.Sprintf("failures error: %v", err)
	}
	if len(deps) == 0 {
		return "failures: no failed deployments"
	}
	var lines []string
	for _, d := range deps {
		logs, _ := r.client.DeploymentLogs(d.ID, 3)
		summary := "no logs"
		if len(logs) > 0 {
			summary = logs[0]
			if len(logs) > 1 {
				summary = fmt.Sprintf("%s, %s...", logs[0], logs[1])
			}
		}
		lines = append(lines, fmt.Sprintf("- %s: %s - %s", d.ID, d.Trigger, summary))
	}
	return "failures:\n" + strings.Join(lines, "\n")
}

func (r *Router) help() string {
	return `Available commands:
/status   - Show running/stopped services
/health   - Show health summary with failed deployments
/recent   - Show latest 5 deployments
/failures - Show recent failed deployments
/help     - Show this help`
}

func (r *Router) freeText(message string) string {
	if r.llm == nil {
		return "llm error: no llm configured"
	}
	resp, err := r.llm.Generate(message)
	if err != nil {
		return fmt.Sprintf("llm error: %v", err)
	}
	return resp
}