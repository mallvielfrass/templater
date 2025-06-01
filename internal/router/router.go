package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Router struct {
	router *chi.Mux
}

func NewRouter() *Router {

	return &Router{

		router: chi.NewRouter(),
	}
}
func (root *Router) Mount() {
	root.router.Use(middleware.Logger)
	//disable cors

	root.router.Route("/api", func(r chi.Router) {

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
