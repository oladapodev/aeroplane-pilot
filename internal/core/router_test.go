package core

import (
	"errors"
	"testing"

	"github.com/oladapodev/aeroplane-pilot/internal/aeroplane"
)

type fakeAeroplane struct {
	services    []aeroplane.Service
	failed      []aeroplane.Deployment
	listErr     error
	failedErr   error
	listCalls   int
	failedCalls int
}

func (f *fakeAeroplane) ListServices() ([]aeroplane.Service, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.services, nil
}

func (f *fakeAeroplane) FailedDeployments(limit int) ([]aeroplane.Deployment, error) {
	f.failedCalls++
	if f.failedErr != nil {
		return nil, f.failedErr
	}
	if limit > 0 && len(f.failed) > limit {
		return f.failed[:limit], nil
	}
	return f.failed, nil
}

type fakeLLM struct {
	prompt   string
	response string
	err      error
	calls    int
}

func (f *fakeLLM) Generate(prompt string) (string, error) {
	f.calls++
	f.prompt = prompt
	if f.err != nil {
		return "", f.err
	}
	return f.response, nil
}

func TestRouteStatus(t *testing.T) {
	router := NewRouter(&fakeAeroplane{
		services: []aeroplane.Service{
			{Name: "api", Status: "running"},
			{Name: "web", Status: "stopped"},
		},
		failed: []aeroplane.Deployment{{ID: "dep-1", Status: "failed"}},
	})

	got := router.Route("/status")
	want := "status: 2 services, 1 running, 1 stopped"
	if got != want {
		t.Fatalf("Route(/status) = %q, want %q", got, want)
	}
}

func TestRouteStatusSurfaceError(t *testing.T) {
	router := NewRouter(&fakeAeroplane{listErr: errors.New("db unavailable")})

	got := router.Route("/status")
	want := "status error: db unavailable"
	if got != want {
		t.Fatalf("Route(/status) = %q, want %q", got, want)
	}
}

func TestRouteHealth(t *testing.T) {
	router := NewRouter(&fakeAeroplane{
		services: []aeroplane.Service{
			{Name: "api", Status: "running"},
			{Name: "web", Status: "failed"},
		},
		failed: []aeroplane.Deployment{
			{ID: "dep-1", Status: "failed"},
			{ID: "dep-2", Status: "failed"},
		},
	})

	got := router.Route("/health")
	want := "health: 2 services, 1 healthy, 1 unhealthy, 2 failed deployments"
	if got != want {
		t.Fatalf("Route(/health) = %q, want %q", got, want)
	}
}

func TestRouteHealthSurfaceError(t *testing.T) {
	router := NewRouter(&fakeAeroplane{failedErr: errors.New("query failed")})

	got := router.Route("/health")
	want := "health error: query failed"
	if got != want {
		t.Fatalf("Route(/health) = %q, want %q", got, want)
	}
}

func TestRouteFreeTextUsesLLM(t *testing.T) {
	llm := &fakeLLM{response: "reply"}
	router := NewRouter(&fakeAeroplane{}, WithLLM(llm))

	got := router.Route("what is the latest deployment status?")
	if got != "reply" {
		t.Fatalf("Route(free text) = %q, want %q", got, "reply")
	}
	if llm.calls != 1 {
		t.Fatalf("llm.calls = %d, want 1", llm.calls)
	}
	wantPrompt := "User message: what is the latest deployment status?\nRespond concisely."
	if llm.prompt != wantPrompt {
		t.Fatalf("llm.prompt = %q, want %q", llm.prompt, wantPrompt)
	}
}

func TestRouteFreeTextSurfaceLLMError(t *testing.T) {
	router := NewRouter(&fakeAeroplane{}, WithLLM(&fakeLLM{err: errors.New("llm down")}))

	got := router.Route("hello")
	want := "llm error: llm down"
	if got != want {
		t.Fatalf("Route(free text) = %q, want %q", got, want)
	}
}
