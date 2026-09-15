package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pedrobarros/vitaltag-api/internal/service"
)

type contextKey string

const UsuarioIDKey contextKey = "usuario_id"

// JWT extrai e valida o Bearer token, injetando o usuarioID no contexto.
func JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			responderErro(w, http.StatusUnauthorized, "token ausente ou mal formado")
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := service.ValidarJWT(tokenStr)
		if err != nil {
			responderErro(w, http.StatusUnauthorized, "token inválido ou expirado")
			return
		}

		ctx := context.WithValue(r.Context(), UsuarioIDKey, claims.UsuarioID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UsuarioIDDoContexto extrai o ID do usuário autenticado.
func UsuarioIDDoContexto(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UsuarioIDKey).(uuid.UUID)
	return id, ok
}

func responderErro(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"erro": msg})
}
