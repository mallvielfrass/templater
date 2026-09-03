package router

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	router        *chi.Mux
	fileStorage   fileStorage
	userStorage   userStorage
	taskStorage   taskStorage
	jwtSecret     string
	publicBaseURL string
	ooJwtSecret   string
	ooInternalURL string
	corsOrigins   []string
	staticDir     string
	saveWaitersMu sync.Mutex
	saveWaiters   map[string][]chan struct{}
}

func NewRouter(fileStorage fileStorage, userStorage userStorage, taskStorage taskStorage, jwtSecret, publicBaseURL, ooJwtSecret, ooInternalURL string) *Router {
	return &Router{
		router:        chi.NewRouter(),
		fileStorage:   fileStorage,
		userStorage:   userStorage,
		taskStorage:   taskStorage,
		jwtSecret:     jwtSecret,
		publicBaseURL: stringsTrimSlash(publicBaseURL),
		ooJwtSecret:   ooJwtSecret,
		ooInternalURL: stringsTrimSlash(ooInternalURL),
		saveWaiters:   make(map[string][]chan struct{}),
	}
}

func (root *Router) SetCORSOrigins(origins []string) {
	root.corsOrigins = origins
}

func (root *Router) SetStaticDir(dir string) {
	root.staticDir = dir
}

func (root *Router) registerSaveWaiter(key string) (chan struct{}, func()) {
	root.saveWaitersMu.Lock()
	defer root.saveWaitersMu.Unlock()
	if root.saveWaiters == nil {
		root.saveWaiters = make(map[string][]chan struct{})
	}
	ch := make(chan struct{}, 1)
	root.saveWaiters[key] = append(root.saveWaiters[key], ch)
	cleanup := func() {
		root.saveWaitersMu.Lock()
		defer root.saveWaitersMu.Unlock()
		list := root.saveWaiters[key]
		for i, c := range list {
			if c == ch {
				root.saveWaiters[key] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(root.saveWaiters[key]) == 0 {
			delete(root.saveWaiters, key)
		}
	}
	return ch, cleanup
}

func (root *Router) notifySaveWaiters(key string) {
	root.saveWaitersMu.Lock()
	defer root.saveWaitersMu.Unlock()
	if root.saveWaiters == nil {
		return
	}
	for _, ch := range root.saveWaiters[key] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func stringsTrimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func (root *Router) Mount() {
	root.router.Use(middleware.Logger)
	root.router.Use(root.CORS)
	root.router.Post("/api/user", root.CreateTempUser)
	root.router.Get("/api/files/{hash}", root.GetFile)
	root.router.Post("/api/onlyoffice/callback", root.OnlyOfficeCallback)
	root.router.Route("/api", func(r chi.Router) {
		r.Use(root.JWTMiddleware)
		r.Post("/create_task", root.CreateTask)
		r.Post("/xlsx_info", root.XlsxInfo)
		r.Post("/run_task", root.RunTask)
		r.Post("/download_zip", root.DownloadZip)
		r.Get("/columns", root.Columns)
		r.Get("/sheet_info", root.SheetInfo)
		r.Get("/onlyoffice/config", root.OnlyOfficeConfig)
	})
	root.router.Get("/onlyoffice-plugin-config.json", root.PluginConfig)
	root.router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	if !root.mountStatic() {
		root.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome"))
		})
	}
}

func (root *Router) mountStatic() bool {
	if root.staticDir == "" {
		return false
	}
	info, err := os.Stat(root.staticDir)
	if err != nil || !info.IsDir() {
		return false
	}
	fileRoot := http.Dir(root.staticDir)
	fileServer := http.FileServer(fileRoot)
	root.router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" || r.URL.Path == "/ping" {
			http.NotFound(w, r)
			return
		}
		fsPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if fsPath == "" {
			fsPath = "index.html"
		}
		f, err := fileRoot.Open(fsPath)
		if err != nil {
			if os.IsNotExist(err) || errors.Is(err, fs.ErrNotExist) {
				r2 := r.Clone(r.Context())
				r2.URL.Path = "/"
				fileServer.ServeHTTP(w, r2)
				return
			}
			http.NotFound(w, r)
			return
		}
		stat, _ := f.Stat()
		_ = f.Close()
		if stat != nil && stat.IsDir() {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	}))
	return true
}

func (r *Router) Run(port int) error {
	fmt.Printf("http server Listening on port %d\n", port)
	return http.ListenAndServe(":"+fmt.Sprint(port), r.router)
}
