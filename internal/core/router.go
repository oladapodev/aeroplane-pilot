package core

import (
	"fmt"
	"strings"

	"github.com/oladapodev/aeroplane-pilot/internal/aeroplane"
	botpkg "github.com/oladapodev/aeroplane-pilot/internal/bot"
	"github.com/oladapodev/aeroplane-pilot/internal/llm"
)

type AeroplaneClient interface {
	ListServices() ([]aeroplane.Service, error)
	FailedDeployments(limit int) ([]aeroplane.Deployment, error)
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

func (r *Router) Handle(message botpkg.Message) string {
	return r.Route(message.Text)
}

func (r *Router) Route(message string) string {
	message = strings.TrimSpace(message)
	switch message {
	case "/status":
		return r.status()
	case "/health":
		return r.health()
	default:
		return r.generate(message)
	}
}

func (r *Router) status() string {
	services, err := r.client.ListServices()
	if err != nil {
		return fmt.Sprintf("status error: %v", err)
	}

	running := 0
	stopped := 0
	for _, service := range services {
		if strings.EqualFold(service.Status, "running") {
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
	failedDeployments, err := r.client.FailedDeployments(10)
	if err != nil {
		return fmt.Sprintf("health error: %v", err)
	}

	healthy := 0
	unhealthy := 0
	for _, service := range services {
		if strings.EqualFold(service.Status, "running") {
			healthy++
		} else {
			unhealthy++
		}
	}

	return fmt.Sprintf("health: %d services, %d healthy, %d unhealthy, %d failed deployments", len(services), healthy, unhealthy, len(failedDeployments))
}

func (r *Router) generate(message string) string {
	if r.llm == nil {
		return "llm error: no llm configured"
	}
	prompt := fmt.Sprintf("User message: %s\nRespond concisely.", message)
	response, err := r.llm.Generate(prompt)
	if err != nil {
		return fmt.Sprintf("llm error: %v", err)
	}
	return response
}
