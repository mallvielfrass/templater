package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const (
	maxOOCallbackBytes = 1 << 20
	maxOODownloadBytes = 100 << 20
	insertColumnPlugin = "asc.{D68097C2-7901-49BD-950E-1647893DA8A1}"
)

var ooDisabledPlugins = []string{
	"asc.{9DC93CDB-B576-4F0C-B55E-FCC9C48DD007}",
	"asc.{b78a062b-e349-4634-8a44-99825600d299}",
	"asc.{38E022EA-AD92-45FC-B22B-49DF39746DB4}",
	"asc.{7327FC95-16DA-41D9-9AF2-0E7F449F6800}",
	"asc.{440EBF13-9B19-4BD8-8621-05200E58140B}",
	"asc.{D71C2EF0-F15B-47C7-80E9-86D671F9C595}",
	"asc.{BE5CBF95-C0AD-4842-B157-AC40FEDD9841}",
	"asc.{BE5CBF95-C0AD-4842-B157-AC40FEDD9441}",
	"asc.{BE5CBF95-C0AD-4842-B157-AC40FEDD9840}",
	"asc.{BFC5D5C6-89DE-4168-9565-ABD8D1E48711}",
	"asc.{07FD8DFA-DFE0-4089-A124-0730933CC80A}",
}

type ooCallback struct {
	Key    string `json:"key"`
	Status int    `json:"status"`
	URL    string `json:"url"`
	Token  string `json:"token"`
}

func (root *Router) fileUser(w http.ResponseWriter, r *http.Request, hash string) (string, bool) {
	if hash == "" {
		http.NotFound(w, r)
		return "", false
	}
	if token := r.URL.Query().Get("token"); token != "" {
		user, err := parseFileToken(root.jwtSecret, token, hash)
		if err != nil {
			http.NotFound(w, r)
			return "", false
		}
		return user, true
	}
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" || tokenString == authHeader {
		http.NotFound(w, r)
		return "", false
	}
	user, err := parseJWT(root.jwtSecret, tokenString)
	if err != nil || user == "" || !root.userStorage.IsUserExist(user) {
		http.NotFound(w, r)
		return "", false
	}
	return user, true
}

func (root *Router) GetFile(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	user, ok := root.fileUser(w, r, hash)
	if !ok {
		return
	}
	owner, err := root.fileStorage.OwnerOf(hash)
	if err != nil || owner != user {
		http.NotFound(w, r)
		return
	}
	data, err := root.fileStorage.GetAnyDocData(hash)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", "inline; filename=\"document.docx\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (root *Router) OnlyOfficeConfig(w http.ResponseWriter, r *http.Request) {
	task, ok := root.ownedTask(r, taskIDFrom(r))
	if !ok {
		http.NotFound(w, r)
		return
	}
	user := userFrom(r)
	owner, err := root.fileStorage.OwnerOf(task.DocHash)
	if err != nil || owner != user {
		http.NotFound(w, r)
		return
	}
	fileToken, err := generateFileToken(root.jwtSecret, user, task.DocHash)
	if err != nil {
		http.Error(w, "Error creating file token", http.StatusInternalServerError)
		return
	}
	key := uuid.New().String()
	if err := root.fileStorage.SaveOOMapping(key, user, task.DocHash, task.ID); err != nil {
		http.Error(w, "Error saving editor key", http.StatusInternalServerError)
		return
	}
	fileURL := fmt.Sprintf("%s/api/files/%s?token=%s", root.publicBaseURL, task.DocHash, url.QueryEscape(fileToken))
	callbackURL := root.publicBaseURL + "/api/onlyoffice/callback"
	pluginConfigURL := root.pluginBaseURL(r) + "/onlyoffice-plugin-config.json"

	document := map[string]any{
		"fileType": "docx",
		"key":      key,
		"title":    "template.docx",
		"url":      fileURL,
		"permissions": map[string]any{
			"edit":     true,
			"download": true,
			"print":    true,
			"chat":     false,
			"comment":  false,
		},
	}
	editorConfig := map[string]any{
		"callbackUrl": callbackURL,
		"mode":        "edit",
		"lang":        "ru",
		"user": map[string]any{
			"id":   user,
			"name": "user",
		},
		"coEditing": map[string]any{
			"mode":   "fast",
			"change": false,
		},
		"customization": map[string]any{
			"about":          false,
			"help":           false,
			"forcesave":      true,
			"autosave":       true,
			"macros":         false,
			"macrosMode":     "disable",
			"suggestFeature": false,
			"anonymous": map[string]any{
				"request": false,
			},
			"feedback": map[string]any{
				"visible": false,
			},
			"features": map[string]any{
				"featuresTips": false,
			},
		},
		"plugins": map[string]any{
			"autostart": []string{insertColumnPlugin},
			"disable":   ooDisabledPlugins,
			"pluginsData": []string{
				pluginConfigURL,
			},
		},
	}
	payload := jwt.MapClaims{
		"document":     document,
		"editorConfig": editorConfig,
	}
	token, err := signMapClaims(root.ooJwtSecret, payload)
	if err != nil {
		http.Error(w, "Error signing editor config", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"config": map[string]any{
			"document":     document,
			"documentType": "word",
			"editorConfig": editorConfig,
			"token":        token,
		},
		"doc_hash": task.DocHash,
		"key":      key,
	})
}

func (root *Router) OnlyOfficeCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxOOCallbackBytes+1))
	if err != nil || int64(len(body)) > maxOOCallbackBytes {
		writeJSON(w, http.StatusOK, map[string]int{"error": 1})
		return
	}
	cb, ok := root.parseOOCallback(r, body)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]int{"error": 1})
		return
	}
	userID, _, taskID, err := root.fileStorage.GetOOMapping(cb.Key)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]int{"error": 1})
		return
	}
	if cb.Status == 2 || cb.Status == 6 {
		if cb.URL == "" {
			writeJSON(w, http.StatusOK, map[string]int{"error": 1})
			return
		}
		data, err := root.downloadOOFile(cb.URL)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]int{"error": 1})
			return
		}
		newHash, err := root.fileStorage.SaveDocFile("template.docx", data)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]int{"error": 1})
			return
		}
		if err := root.fileStorage.BindOwner(newHash, userID); err != nil {
			writeJSON(w, http.StatusOK, map[string]int{"error": 1})
			return
		}
		if err := root.taskStorage.AddDocHash(taskID, newHash); err != nil {
			writeJSON(w, http.StatusOK, map[string]int{"error": 1})
			return
		}
		_ = root.fileStorage.SaveOOMapping(cb.Key, userID, newHash, taskID)
		root.notifySaveWaiters(cb.Key)
	}
	writeJSON(w, http.StatusOK, map[string]int{"error": 0})
}

func (root *Router) PluginConfig(w http.ResponseWriter, r *http.Request) {
	base := root.pluginBaseURL(r) + "/onlyoffice-plugin/"
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "InsertColumn",
		"guid":    insertColumnPlugin,
		"baseUrl": base,
		"variations": []map[string]any{
			{
				"description":    "Insert Column Plugin",
				"url":            "index.html",
				"icons":          []string{"icon.png", "icon@2x.png"},
				"isViewer":       false,
				"EditorsSupport": []string{"word"},
				"isVisual":       false,
				"isSystem":       false,
				"initDataType":   "none",
				"initData":       "",
				"buttons":        []any{},
			},
		},
	})
}

func requestPublicBase(r *http.Request) string {
	proto := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0])
	host := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0])
	if host == "" {
		host = r.Host
	}
	if host == "" {
		return ""
	}
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	return stringsTrimSlash(proto + "://" + host)
}

func isDockerInternalBase(base string) bool {
	u, err := url.Parse(base)
	if err != nil {
		return true
	}
	h := strings.ToLower(u.Hostname())
	return h == "templater-backend" || h == "onlyoffice" || h == "nginx" ||
		h == "localhost" || h == "127.0.0.1" || strings.HasSuffix(h, ".internal")
}

func (root *Router) pluginBaseURL(r *http.Request) string {
	if b := requestPublicBase(r); b != "" && !isDockerInternalBase(b) {
		return b
	}
	return root.publicBaseURL
}

func (root *Router) parseOOCallback(r *http.Request, body []byte) (ooCallback, bool) {
	if root.ooJwtSecret == "" {
		return ooCallback{}, false
	}
	var cb ooCallback
	_ = json.Unmarshal(body, &cb)
	token := cb.Token
	auth := r.Header.Get("Authorization")
	if bearer := strings.TrimPrefix(auth, "Bearer "); bearer != auth && bearer != "" {
		token = bearer
	}
	if token == "" {
		return ooCallback{}, false
	}
	claims, err := parseMapClaims(root.ooJwtSecret, token)
	if err != nil {
		return ooCallback{}, false
	}
	if !applyOOClaims(&cb, claims) {
		return ooCallback{}, false
	}
	return cb, true
}

func applyOOClaims(cb *ooCallback, claims jwt.MapClaims) bool {
	payload := jwt.MapClaims(claims)
	if nested, ok := claims["payload"].(map[string]any); ok {
		payload = jwt.MapClaims(nested)
	}
	if key, ok := payload["key"].(string); ok && key != "" {
		cb.Key = key
	}
	if u, ok := payload["url"].(string); ok {
		cb.URL = u
	}
	switch v := payload["status"].(type) {
	case float64:
		cb.Status = int(v)
	case json.Number:
		n, _ := v.Int64()
		cb.Status = int(n)
	}
	return cb.Key != ""
}

func ooDownloadPathAllowed(rawPath string) bool {
	clean := path.Clean(rawPath)
	if clean == "." || clean == "/" {
		return false
	}
	lower := strings.ToLower(clean)
	return strings.HasPrefix(lower, "/cache/") || strings.HasPrefix(lower, "/download")
}

func (root *Router) downloadOOFile(rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid scheme")
	}
	if !ooDownloadPathAllowed(u.Path) {
		return nil, fmt.Errorf("invalid path")
	}
	if root.ooInternalURL == "" {
		return nil, fmt.Errorf("onlyoffice url is not set")
	}
	base, err := url.Parse(root.ooInternalURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = base.Scheme
	u.Host = base.Host
	u.User = nil

	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxOODownloadBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxOODownloadBytes {
		return nil, fmt.Errorf("file too large")
	}
	return data, nil
}

func (root *Router) forceSaveTaskDocument(taskID string) error {
	if root.ooInternalURL == "" {
		return nil
	}
	key, err := root.fileStorage.GetTaskOOKey(taskID)
	if err != nil || key == "" {
		return nil
	}

	ch, cleanup := root.registerSaveWaiter(key)
	defer cleanup()

	cmdBody := map[string]any{
		"c":   "forcesave",
		"key": key,
	}
	if root.ooJwtSecret != "" {
		token, err := signMapClaims(root.ooJwtSecret, cmdBody)
		if err == nil {
			cmdBody["token"] = token
		}
	}

	data, err := json.Marshal(cmdBody)
	if err != nil {
		return err
	}

	cmdURL := root.ooInternalURL + "/command"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(cmdURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var cmdResp struct {
		Error int `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cmdResp); err != nil {
		return nil
	}

	if cmdResp.Error == 0 {
		select {
		case <-ch:
		case <-time.After(3 * time.Second):
		}
	}

	return nil
}
