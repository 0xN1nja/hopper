package main

import (
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/user/hopper/internal/auth"
	"github.com/user/hopper/internal/config"
	"github.com/user/hopper/internal/db"
	"github.com/user/hopper/internal/handlers"
	authmw "github.com/user/hopper/internal/middleware"
	"github.com/user/hopper/internal/rcon"
	hopperStatic "github.com/user/hopper/static"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DataPath + "/hopper.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := auth.SeedUser(database, cfg.Username, cfg.Password); err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	manager := rcon.NewManager()
	authHandler := auth.NewHandler(database)
	sessionsHandler := handlers.NewSessionsHandler(database, manager)
	rconHandler := handlers.NewRconHandler(database, manager)
	iconHandler := handlers.NewIconHandler(cfg.DataPath)
	authMiddleware := authmw.Auth(database)

	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	if cfg.Dev {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:5173"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Cookie"},
			AllowCredentials: true,
		}))
	}

	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/logout", authHandler.Logout)
	r.With(authMiddleware).Get("/api/auth/me", authHandler.Me)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/api/sessions", sessionsHandler.List)
		r.Post("/api/sessions", sessionsHandler.Create)
		r.Put("/api/sessions/{id}", sessionsHandler.Update)
		r.Delete("/api/sessions/{id}", sessionsHandler.Delete)
		r.Get("/api/sessions/{id}/history", sessionsHandler.GetHistory)
		r.Delete("/api/sessions/{id}/history", sessionsHandler.ClearHistory)

		r.Post("/api/sessions/{id}/test", rconHandler.Test)
		r.Post("/api/sessions/{id}/connect", rconHandler.Connect)
		r.Post("/api/sessions/{id}/disconnect", rconHandler.Disconnect)
		r.Get("/api/sessions/{id}/status", rconHandler.Status)
		r.Get("/api/sessions/{id}/ws", rconHandler.WebSocket)

		r.Post("/api/icons", iconHandler.Upload)
	})

	r.Get("/api/icons/{filename}", iconHandler.Serve)

	r.Handle("/*", spaHandler())

	log.Printf("hopper listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func spaHandler() http.Handler {
	fsys, err := fs.Sub(hopperStatic.FS, "files")
	if err != nil {
		log.Fatalf("failed to create sub FS: %v", err)
	}
	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		_, err := fsys.Open(path)
		if err != nil {
			r2 := r.Clone(r.Context())
			r2.URL = &url.URL{Path: "/"}
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
