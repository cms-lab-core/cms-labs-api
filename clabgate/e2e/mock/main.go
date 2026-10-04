package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	attemptID = "00000000-0000-0000-0000-000000000000"
	revision  = "0123456789abcdef0123456789abcdef01234567"
)

type mockState struct {
	sync.RWMutex
	Status string `json:"status"`
	Result any    `json:"result,omitempty"`
}

func main() {
	state := &mockState{Status: "pending"}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	stateHandler := func(w http.ResponseWriter, _ *http.Request) {
		state.RLock()
		defer state.RUnlock()
		writeJSON(w, state)
	}
	mux.HandleFunc("GET /state", stateHandler)
	mux.HandleFunc("GET /api/state", stateHandler)
	mux.HandleFunc("GET /api/v1/sso/userinfo", handleUserInfo)
	mux.HandleFunc("POST /api/v1/rpc", func(w http.ResponseWriter, r *http.Request) {
		handleRPC(w, r, state)
	})
	mux.HandleFunc("GET /", handleGitLab)

	server := &http.Server{
		Addr: ":8080", Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second,
	}
	log.Printf("clabgate smoke mock listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]any{
		"sub": "42", "email": "student@example.com", "name": "smoke-student",
		"username": "smoke-student", "roles": []string{"student"},
	})
}

func handleRPC(w http.ResponseWriter, r *http.Request, state *mockState) {
	username, password, ok := r.BasicAuth()
	if !ok || username != "k8s" || password != "smoke-secret" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var request struct {
		ID     any             `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch request.Method {
	case "lti_attempt.list_external":
		state.RLock()
		status := state.Status
		result := state.Result
		state.RUnlock()
		writeJSON(w, map[string]any{
			"jsonrpc": "2.0", "id": request.ID,
			"result": map[string]any{"model": []map[string]any{{
				"id": 1, "attempt_id": attemptID, "user_id": 42, "user_name": "smoke-student",
				"status": status, "server_client_id": "k8s", "lti_routing_name": "Local smoke lab",
				"repository": "", "test_path": "", "result": result,
			}}},
		})
	case "lti_attempt.update_external":
		var params struct {
			Models []struct {
				AttemptID string `json:"attempt_id"`
				Status    string `json:"status"`
				Result    any    `json:"result"`
			} `json:"models"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		count := 0
		state.Lock()
		for _, model := range params.Models {
			if model.AttemptID != attemptID {
				continue
			}
			if model.Status != "" {
				state.Status = model.Status
			}
			if model.Result != nil {
				state.Result = model.Result
			}
			count++
		}
		state.Unlock()
		writeJSON(w, map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"count": count}})
	default:
		writeJSON(w, map[string]any{
			"jsonrpc": "2.0", "id": request.ID,
			"error": map[string]any{"code": -32601, "message": "method not found"},
		})
	}
}

func handleGitLab(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.EscapedPath()
	switch {
	case strings.Contains(requestPath, "/repository/commits/master"):
		writeJSON(w, map[string]string{"id": revision})
	case strings.HasSuffix(requestPath, "/repository/tree"):
		writeJSON(w, []map[string]string{{"path": "topology.template.yaml", "type": "blob"}})
	case strings.Contains(requestPath, "/repository/files/") && strings.HasSuffix(requestPath, "/raw"):
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write([]byte(`apiVersion: c9s.run/v1alpha1
kind: Topology
metadata:
  name: smoke
spec:
  definition:
    containerlab: |
      name: smoke
      topology:
        nodes:
          client:
            kind: linux
            image: docker.io/library/alpine:3.24
            ttyd-shell: /bin/sh
`))
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
