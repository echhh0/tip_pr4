package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/echhh0/tip_pr4/internal/task"
	middleware "github.com/echhh0/tip_pr4/pkg/middleware"
)

func main() {
	repo := task.NewRepo().WithFile("tasks.json")
	if err := repo.Load(); err != nil {
		log.Fatalf("load tasks: %v", err)
	}
	h := task.NewHandler(repo)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.SimpleCORS)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/api", func(api chi.Router) {
		api.Route("/v1", func(v1 chi.Router) {
			v1.Mount("/tasks", h.Routes())
		})
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
