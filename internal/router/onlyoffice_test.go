package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/require"

	"github.com/mallvielfrass/templater/internal/models"
)

func TestOOCallbackRejectsEmptySecret(t *testing.T) {
	st := &testStorage{oo: map[string][3]string{"k1": {"u1", "h1", "t1"}}}
	r := NewRouter(st, nil, nil, "jwt", "http://base", "", "http://oo")
	r.Mount()

	body, _ := json.Marshal(ooCallback{Key: "k1", Status: 2, URL: "http://evil/cache/x"})
	req := httptest.NewRequest(http.MethodPost, "/api/onlyoffice/callback", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)

	var resp map[string]int
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, 1, resp["error"])
}

func TestOOCallbackUsesJWTURLNotBody(t *testing.T) {
	payload := []byte("from-jwt")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.True(t, strings.HasPrefix(r.URL.Path, "/cache/"))
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	st := &testStorage{
		oo:     map[string][3]string{"doc-key": {"user-1", "old-hash", "task-1"}},
		docs:   map[string][]byte{},
		owners: map[string]string{},
	}
	tasks := &testTaskStorage{tasks: map[string]models.Task{
		"task-1": {ID: "task-1", UserID: "user-1", DocHash: "old-hash"},
	}}
	r := NewRouter(st, nil, tasks, "jwt", "http://base", "oo-secret", ts.URL)
	r.Mount()

	claims := jwt.MapClaims{
		"key":    "doc-key",
		"status": float64(2),
		"url":    "http://onlyoffice/cache/files/ok.docx",
	}
	token, err := signMapClaims("oo-secret", claims)
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]any{
		"key":    "doc-key",
		"status": 2,
		"url":    "http://evil.example/cache/pwn.docx",
		"token":  token,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/onlyoffice/callback", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)

	var resp map[string]int
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, 0, resp["error"])
	require.Equal(t, payload, st.docs[tasks.tasks["task-1"].DocHash])
}

func TestOOCallbackRejectsCommandPath(t *testing.T) {
	st := &testStorage{oo: map[string][3]string{"doc-key": {"user-1", "h", "task-1"}}}
	r := NewRouter(st, nil, &testTaskStorage{tasks: map[string]models.Task{
		"task-1": {ID: "task-1", UserID: "user-1"},
	}}, "jwt", "http://base", "oo-secret", "http://127.0.0.1:9")
	r.Mount()

	claims := jwt.MapClaims{
		"key":    "doc-key",
		"status": float64(2),
		"url":    "http://onlyoffice/command",
	}
	token, err := signMapClaims("oo-secret", claims)
	require.NoError(t, err)
	body, _ := json.Marshal(ooCallback{Key: "doc-key", Status: 2, URL: "http://onlyoffice/command", Token: token})
	req := httptest.NewRequest(http.MethodPost, "/api/onlyoffice/callback", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	var resp map[string]int
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, 1, resp["error"])
}

func TestOOConfigIgnoresPluginBaseURL(t *testing.T) {
	st := &testStorage{
		owners: map[string]string{"doc-hash": "user-1"},
		docs:   map[string][]byte{"doc-hash": {1, 2, 3}},
	}
	us := &testUserStorage{users: map[string]bool{"user-1": true}}
	tasks := &testTaskStorage{tasks: map[string]models.Task{
		"task-1": {ID: "task-1", UserID: "user-1", DocHash: "doc-hash"},
	}}
	r := NewRouter(st, us, tasks, "jwt-secret", "https://app.example", "oo-secret", "http://onlyoffice")
	r.Mount()

	token := generateJWT("jwt-secret", "user-1")
	req := httptest.NewRequest(http.MethodGet, "/api/onlyoffice/config?task_id=task-1&plugin_base_url=https://evil.example", nil)
	req.Host = "app.example"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.NotContains(t, rr.Body.String(), "evil.example")
	require.Contains(t, rr.Body.String(), "https://app.example/onlyoffice-plugin/config.json")
	require.Contains(t, rr.Body.String(), `"chat":false`)
}

func TestCORSAllowlist(t *testing.T) {
	r := NewRouter(&testStorage{}, &testUserStorage{users: map[string]bool{}}, nil, "jwt", "http://base", "oo", "http://oo")
	r.SetCORSOrigins([]string{"http://localhost:5173"})
	r.Mount()

	req := httptest.NewRequest(http.MethodOptions, "/api/user", nil)
	req.Header.Set("Origin", "http://evil.example")
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, "", rr.Header().Get("Access-Control-Allow-Origin"))

	req = httptest.NewRequest(http.MethodOptions, "/api/user", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr = httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, "http://localhost:5173", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestDownloadPathTraversalRejected(t *testing.T) {
	require.False(t, ooDownloadPathAllowed("/cache/../../command"))
	require.False(t, ooDownloadPathAllowed("/command"))
	require.True(t, ooDownloadPathAllowed("/cache/files/a.docx"))
	require.True(t, ooDownloadPathAllowed("/download"))
}
