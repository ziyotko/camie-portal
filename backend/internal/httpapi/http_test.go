package httpapi

import (
	"camie-portal/backend/internal/generator"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testPublisher struct{ done chan struct{} }

func (p testPublisher) GenerateSite(ctx context.Context) (generator.Result, error) {
	select {
	case <-p.done:
		return generator.Result{Files: 5}, nil
	case <-ctx.Done():
		return generator.Result{}, ctx.Err()
	}
}
func TestAuthenticatedAsyncPublication(t *testing.T) {
	done := make(chan struct{})
	token := strings.Repeat("x", 32)
	handler := New(context.Background(), testPublisher{done}, token, time.Second)
	request := func(method, path string, auth bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		if auth {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	if w := request("GET", "/healthz", false); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := request("POST", "/api/static/site", false); w.Code != 401 {
		t.Fatal(w.Code)
	}
	raw := httptest.NewRequest("POST", "/api/static/site", nil)
	raw.Header.Set("Authorization", token)
	rawResponse := httptest.NewRecorder()
	handler.ServeHTTP(rawResponse, raw)
	if rawResponse.Code != 401 {
		t.Fatal("missing Bearer scheme was accepted")
	}
	if w := request("GET", "/api/static/site", true); w.Code != http.StatusMethodNotAllowed {
		t.Fatal(w.Code)
	}
	w := request("POST", "/api/static/site", true)
	if w.Code != 202 {
		t.Fatal(w.Code)
	}
	var job Job
	if err := json.Unmarshal(w.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	if w = request("POST", "/api/static/site", true); w.Code != 409 {
		t.Fatal("concurrent job accepted")
	}
	close(done)
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		w = request("GET", "/api/static/jobs/"+job.ID, true)
		if err := json.Unmarshal(w.Body.Bytes(), &job); err != nil {
			t.Fatal(err)
		}
		if job.Status == "completed" {
			if job.Result.Files != 5 {
				t.Fatal(job)
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("job did not complete")
}
