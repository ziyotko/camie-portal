package httpapi

import (
	"camie-portal/backend/internal/generator"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Publisher interface {
	GenerateSite(context.Context) (generator.Result, error)
}
type Job struct {
	ID         string            `json:"id"`
	Status     string            `json:"status"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt *time.Time        `json:"finished_at,omitempty"`
	Result     *generator.Result `json:"result,omitempty"`
	Error      string            `json:"error,omitempty"`
}
type API struct {
	publisher Publisher
	token     string
	ctx       context.Context
	timeout   time.Duration
	mu        sync.Mutex
	jobs      map[string]Job
	running   bool
}

func New(ctx context.Context, p Publisher, token string, timeout time.Duration) http.Handler {
	return &API{publisher: p, token: token, ctx: ctx, timeout: timeout, jobs: map[string]Job{}}
}
func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" && r.Method == http.MethodGet {
		reply(w, 200, map[string]bool{"ok": true})
		return
	}
	authorization := r.Header.Get("Authorization")
	supplied := strings.TrimPrefix(authorization, "Bearer ")
	if !strings.HasPrefix(authorization, "Bearer ") || a.token == "" || subtle.ConstantTimeCompare([]byte(supplied), []byte(a.token)) != 1 {
		reply(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	if r.URL.Path == "/api/static/site" {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			reply(w, 405, map[string]string{"error": "method not allowed"})
			return
		}
		a.mu.Lock()
		if a.running {
			a.mu.Unlock()
			reply(w, 409, map[string]string{"error": "a publish job is already running"})
			return
		}
		// The API retains at most 100 completed job summaries in memory.
		if len(a.jobs) >= 100 {
			var oldest Job
			for _, job := range a.jobs {
				if oldest.ID == "" || job.StartedAt.Before(oldest.StartedAt) {
					oldest = job
				}
			}
			delete(a.jobs, oldest.ID)
		}
		var idBytes [16]byte
		if _, err := rand.Read(idBytes[:]); err != nil {
			a.mu.Unlock()
			reply(w, 500, map[string]string{"error": "cannot create job"})
			return
		}
		id := hex.EncodeToString(idBytes[:])
		job := Job{ID: id, Status: "running", StartedAt: time.Now()}
		a.jobs[id] = job
		a.running = true
		a.mu.Unlock()
		go func() {
			ctx, cancel := context.WithTimeout(a.ctx, a.timeout)
			defer cancel()
			result, err := a.publisher.GenerateSite(ctx)
			a.mu.Lock()
			defer a.mu.Unlock()
			finished := time.Now()
			job.FinishedAt = &finished
			if err != nil {
				job.Status = "failed"
				job.Error = err.Error()
			} else {
				job.Status = "completed"
				job.Result = &result
			}
			a.jobs[id] = job
			a.running = false
		}()
		w.Header().Set("Location", "/api/static/jobs/"+id)
		reply(w, 202, job)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/static/jobs/") && r.Method == http.MethodGet {
		id := strings.TrimPrefix(r.URL.Path, "/api/static/jobs/")
		a.mu.Lock()
		job, exists := a.jobs[id]
		a.mu.Unlock()
		if !exists {
			reply(w, 404, map[string]string{"error": "job not found"})
			return
		}
		reply(w, 200, job)
		return
	}
	reply(w, 404, map[string]string{"error": "not found"})
}
