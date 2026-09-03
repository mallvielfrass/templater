package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	router      *chi.Mux
	fileStorage fileStorage
	userStorage userStorage
	taskStorage taskStorage
	jwtSecret   string
}

func NewRouter(fileStorage fileStorage, userStorage userStorage, taskStorage taskStorage, jwtSecret string) *Router {
	return &Router{
		router:      chi.NewRouter(),
		fileStorage: fileStorage,
		userStorage: userStorage,
		taskStorage: taskStorage,
		jwtSecret:   jwtSecret,
	}
}
func (root *Router) Mount() {
	root.router.Use(middleware.Logger)
	//disable cors
	root.router.Post("/api/user", root.CreateTempUser)
	root.router.Route("/api", func(r chi.Router) {
		r.Use(root.JWTMiddleware)
		r.Post("/create_task", root.CreateTask)

	})
	root.router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	root.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
}
func (r *Router) Run(port int) error {
	fmt.Printf("http server Listening on port %d\n", port)
	return http.ListenAndServe(":"+fmt.Sprint(port), r.router)
}
