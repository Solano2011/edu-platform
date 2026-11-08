package middleware

import (
	"context"
	"net/http"

	"edu-platform/internal/domain/user"
	authUC "edu-platform/internal/usecase/auth"
)

type contextKey string

const UserContextKey contextKey = "authenticated_user"

func AuthMiddleware(authUseCase *authUC.UseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err == nil && cookie.Value != "" {
				u, err := authUseCase.Authenticate(r.Context(), cookie.Value)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserContextKey, u)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(UserContextKey)
		if val == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		u, ok := val.(*user.User)
		if !ok || !u.IsAdmin() {
			http.Error(w, "Доступ запрещен: требуются права администратора", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func UserFromContext(ctx context.Context) *user.User {
	if val := ctx.Value(UserContextKey); val != nil {
		if u, ok := val.(*user.User); ok {
			return u
		}
	}
	return nil
}
