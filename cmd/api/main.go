package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/pedrobarros/vitaltag-api/internal/handler"
	"github.com/pedrobarros/vitaltag-api/internal/middleware"
	"github.com/pedrobarros/vitaltag-api/internal/repository"
	"github.com/pedrobarros/vitaltag-api/internal/service"
)

func main() {
	// Load .env in development (ignored in production when variables are already set)
	_ = godotenv.Load()

	ctx := context.Background()

	// Connection pool
	pool, err := repository.NewPool(ctx)
	if err != nil {
		log.Fatalf("banco de dados: %v", err)
	}
	defer pool.Close()
	log.Println("✅  Conectado ao PostgreSQL")

	// Repositories
	usuarioRepo := repository.NewUsuarioRepo(pool)
	perfilRepo := repository.NewPerfilRepo(pool)

	// Services
	authSvc := service.NewAuthService(usuarioRepo)
	perfilSvc := service.NewPerfilService(perfilRepo, usuarioRepo)
	publicoSvc := service.NewPublicoService(perfilRepo)
	qrCodeSvc := service.NewQRCodeService(perfilRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	perfilH := handler.NewPerfilHandler(perfilSvc, qrCodeSvc)
	publicoH := handler.NewPublicoHandler(publicoSvc)

	// Roteador
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(corsMiddleware)

	// Public routes
	r.Post("/api/v1/auth/registro", authH.Registro)
	r.Post("/api/v1/auth/login", authH.Login)
	r.Post("/api/v1/publico/{slug}", publicoH.AcessarPerfil)

	// Public profile GET (allows the frontend to load without a password)
	r.Get("/api/v1/publico/{slug}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"erro":"use POST com a senha"}`, http.StatusMethodNotAllowed)
	})

	// Rotas protegidas por JWT
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWT)
		r.Get("/api/v1/perfil", perfilH.Buscar)
		r.Post("/api/v1/perfil", perfilH.SalvarOuAtualizar)
		r.Get("/api/v1/perfil/qrcode", perfilH.QRCode)
		r.Delete("/api/v1/perfil", perfilH.Excluir)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀  VitalTag API em http://localhost:%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("servidor: %v", err)
	}
}

// corsMiddleware adiciona os headers de CORS para o frontend Next.js.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("CORS_ORIGIN")
		if origin == "" {
			origin = "http://localhost:3000"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
