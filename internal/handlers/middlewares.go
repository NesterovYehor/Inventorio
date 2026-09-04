package handlers

import (
	"context"
	"net/http"
)

func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "ru"

		if cookie, err := r.Cookie("lang"); err == nil {
			lang = cookie.Value
		}

		if q := r.URL.Query().Get("lang"); q != "" {
			lang = q
			http.SetCookie(w, &http.Cookie{
				Name:     "lang",
				Value:    q,
				Path:     "/",
				MaxAge:   60 * 60 * 24 * 365,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
		}

		// Put the language into the request's "backpack" (Context)
		ctx := context.WithValue(r.Context(), "lang", lang)

		// Pass the request down the chain to your handlers
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
