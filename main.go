package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Method string `json:"method"`
}

type HATEOASResponse struct {
	Data  interface{} `json:"data"`
	Links []Link      `json:"links"`
}

type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	functions   []Function
	functionsMu sync.RWMutex
	tokens      = make(map[string]time.Time)
	tokensMu    sync.Mutex
	functionsDir string
)

func main() {
	functionsDir = "functions"
	os.MkdirAll(functionsDir, 0755)
	scanFunctions()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/functions/", functionHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	go func() {
		for {
			time.Sleep(30 * time.Second)
			cleanTokens()
		}
	}()

	http.ListenAndServe(":8087", nil)
}

func scanFunctions() {
	functionsMu.Lock()
	defer functionsMu.Unlock()
	functions = nil
	entries, err := os.ReadDir(functionsDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			functions = append(functions, Function{
				Name:        e.Name(),
				Description: "Executes " + e.Name(),
				Endpoint:    "/functions/" + e.Name(),
			})
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Header.Get("Accept") == "application/json" {
		functionsMu.RLock()
		defer functionsMu.RUnlock()
		links := []Link{
			{Rel: "auth", Href: "/auth", Method: "POST"},
		}
		for _, f := range functions {
			links = append(links, Link{Rel: "function", Href: f.Endpoint, Method: "POST"})
		}
		resp := HATEOASResponse{Data: functions, Links: links}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	functionsMu.RLock()
	defer functionsMu.RUnlock()
	tmpl.Execute(w, functions)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, HATEOASResponse{
			Links: []Link{{Rel: "auth", Href: "/auth", Method: "POST"}},
		})
		return
	}
	var req AuthRequest
	json.NewDecoder(r.Body).Decode(&req)
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, HATEOASResponse{
			Links: []Link{{Rel: "auth", Href: "/auth", Method: "POST"}},
		})
		return
	}
	h := sha256.Sum256([]byte(req.Username + ":" + req.Password))
	token := hex.EncodeToString(h[:])
	tokensMu.Lock()
	tokens[token] = time.Now().Add(1 * time.Hour)
	tokensMu.Unlock()

	functionsMu.RLock()
	var links []Link
	for _, f := range functions {
		links = append(links, Link{Rel: "function", Href: f.Endpoint, Method: "POST"})
	}
	functionsMu.RUnlock()
	links = append(links, Link{Rel: "self", Href: "/", Method: "GET"})

	writeJSON(w, http.StatusOK, HATEOASResponse{
		Data:  map[string]string{"token": token},
		Links: links,
	})
}

func functionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, HATEOASResponse{
			Links: []Link{{Rel: "auth", Href: "/auth", Method: "POST"}},
		})
		return
	}
	token := r.Header.Get("Authorization")
	tokensMu.Lock()
	expiry, ok := tokens[token]
	tokensMu.Unlock()
	if !ok || time.Now().After(expiry) {
		writeJSON(w, http.StatusUnauthorized, HATEOASResponse{
			Links: []Link{{Rel: "auth", Href: "/auth", Method: "POST"}},
		})
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/functions/")
	if name == "" || strings.Contains(name, "/") {
		writeJSON(w, http.StatusNotFound, HATEOASResponse{
			Links: []Link{{Rel: "functions", Href: "/", Method: "GET"}},
		})
		return
	}

	funcPath := filepath.Join(functionsDir, name)
	if _, err := os.Stat(funcPath); os.IsNotExist(err) {
		writeJSON(w, http.StatusNotFound, HATEOASResponse{
			Links: []Link{{Rel: "functions", Href: "/", Method: "GET"}},
		})
		return
	}

	body, _ := io.ReadAll(r.Body)
	cmd := exec.Command(funcPath)
	cmd.Stdin = strings.NewReader(string(body))
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, HATEOASResponse{
			Data:  map[string]string{"error": err.Error()},
			Links: []Link{{Rel: "functions", Href: "/", Method: "GET"}},
		})
		return
	}

	functionsMu.RLock()
	var links []Link
	for _, f := range functions {
		links = append(links, Link{Rel: "function", Href: f.Endpoint, Method: "POST"})
	}
	functionsMu.RUnlock()

	writeJSON(w, http.StatusOK, HATEOASResponse{
		Data:  string(output),
		Links: links,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func cleanTokens() {
	tokensMu.Lock()
	defer tokensMu.Unlock()
	now := time.Now()
	for k, v := range tokens {
		if now.After(v) {
			delete(tokens, k)
		}
	}
}
