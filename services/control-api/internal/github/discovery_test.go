package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// staticTokenProvider always returns the same token.
type staticTokenProvider struct {
	token string
}

func (s *staticTokenProvider) InstallationToken(_ context.Context) (string, error) {
	return s.token, nil
}

func TestListAccessibleRepos_SinglePage(t *testing.T) {
	repos := []GitHubRepo{
		{ID: 1, FullName: "org/repo1", Name: "repo1"},
		{ID: 2, FullName: "org/repo2", Name: "repo2"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/installation/repositories" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		resp := installationReposResponse{
			TotalCount:   len(repos),
			Repositories: repos,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "test-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	got, err := ds.ListAccessibleRepos(context.Background())
	if err != nil {
		t.Fatalf("ListAccessibleRepos() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d repos, want 2", len(got))
	}
	if got[0].FullName != "org/repo1" {
		t.Errorf("got[0].FullName = %q, want %q", got[0].FullName, "org/repo1")
	}
	if got[1].FullName != "org/repo2" {
		t.Errorf("got[1].FullName = %q, want %q", got[1].FullName, "org/repo2")
	}
}

func TestListAccessibleRepos_Pagination(t *testing.T) {
	// Simulate 150 repos across 2 pages (100 + 50).
	allRepos := make([]GitHubRepo, 150)
	for i := range allRepos {
		allRepos[i] = GitHubRepo{
			ID:       int64(i + 1),
			FullName: fmt.Sprintf("org/repo-%d", i+1),
			Name:     fmt.Sprintf("repo-%d", i+1),
		}
	}

	var requestCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

		if perPage != 100 {
			t.Errorf("per_page = %d, want 100", perPage)
		}

		start := (page - 1) * perPage
		end := start + perPage
		if end > len(allRepos) {
			end = len(allRepos)
		}

		pageRepos := allRepos[start:end]
		resp := installationReposResponse{
			TotalCount:   len(allRepos),
			Repositories: pageRepos,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "test-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	got, err := ds.ListAccessibleRepos(context.Background())
	if err != nil {
		t.Fatalf("ListAccessibleRepos() error = %v", err)
	}

	if len(got) != 150 {
		t.Fatalf("got %d repos, want 150", len(got))
	}
	if requestCount != 2 {
		t.Errorf("made %d requests, want 2", requestCount)
	}
}

func TestListAccessibleRepos_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := installationReposResponse{
			TotalCount:   0,
			Repositories: []GitHubRepo{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "test-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	got, err := ds.ListAccessibleRepos(context.Background())
	if err != nil {
		t.Fatalf("ListAccessibleRepos() error = %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("got %d repos, want 0", len(got))
	}
}

func TestListAccessibleRepos_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "bad-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	_, err := ds.ListAccessibleRepos(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListOrgRepos_SinglePage(t *testing.T) {
	repos := []GitHubRepo{
		{ID: 10, FullName: "myorg/app", Name: "app", Owner: struct {
			Login string `json:"login"`
		}{Login: "myorg"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orgs/myorg/repos" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(repos)
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "test-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	got, err := ds.ListOrgRepos(context.Background(), "myorg")
	if err != nil {
		t.Fatalf("ListOrgRepos() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("got %d repos, want 1", len(got))
	}
	if got[0].FullName != "myorg/app" {
		t.Errorf("got[0].FullName = %q, want %q", got[0].FullName, "myorg/app")
	}
}

func TestListOrgRepos_Pagination(t *testing.T) {
	allRepos := make([]GitHubRepo, 120)
	for i := range allRepos {
		allRepos[i] = GitHubRepo{
			ID:       int64(i + 1),
			FullName: fmt.Sprintf("myorg/repo-%d", i+1),
			Name:     fmt.Sprintf("repo-%d", i+1),
		}
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

		start := (page - 1) * perPage
		end := start + perPage
		if end > len(allRepos) {
			end = len(allRepos)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(allRepos[start:end])
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "test-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	got, err := ds.ListOrgRepos(context.Background(), "myorg")
	if err != nil {
		t.Fatalf("ListOrgRepos() error = %v", err)
	}

	if len(got) != 120 {
		t.Fatalf("got %d repos, want 120", len(got))
	}
}

func TestListOrgRepos_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "test-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	got, err := ds.ListOrgRepos(context.Background(), "empty-org")
	if err != nil {
		t.Fatalf("ListOrgRepos() error = %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("got %d repos, want 0", len(got))
	}
}

func TestListOrgRepos_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "test-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	_, err := ds.ListOrgRepos(context.Background(), "nonexistent-org")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDiscoveryService_AuthHeaderInjected(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		resp := installationReposResponse{TotalCount: 0, Repositories: []GitHubRepo{}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewClient(&staticTokenProvider{token: "my-secret-token"})
	ds := NewDiscoveryService(client, WithDiscoveryBaseURL(srv.URL))

	_, _ = ds.ListAccessibleRepos(context.Background())

	if gotAuth != "token my-secret-token" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "token my-secret-token")
	}
}
