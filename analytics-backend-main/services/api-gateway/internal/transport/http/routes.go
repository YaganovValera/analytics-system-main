package http

import (
	"net/http"

	"github.com/YaganovValera/analytics-system/services/api-gateway/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Routes возвращает основной маршрутизатор с подключёнными роутами и middleware.
func Routes(h *handler.Handler, m *Middleware, jwtMw func(http.Handler) http.Handler, rbacMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	// Встроенные middleware (глобальные)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/v1", func(r chi.Router) {
		// Публичные эндпоинты (без JWT)
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
		r.Post("/refresh", h.Refresh)

		// Группа с JWT + RBAC + обогащением контекста
		r.Group(func(r chi.Router) {
			r.Use(m.WithContext) // Добавляет metadata в gRPC context
			r.Use(jwtMw)         // JWT валидация и context
			r.Use(rbacMw)        // Проверка ролей на основе route

			r.Post("/logout", h.Logout)
			r.Get("/me", h.Me)

			r.Get("/candles", h.GetCandles)
			r.Get("/symbols", h.GetSymbols)
			r.Get("/orderbook", h.GetOrderBook)
			r.Post("/analyze-csv", h.AnalyzeCSV)

			r.Get("/admin/users", h.ListUsers)
			r.Get("/admin/users/{id}", h.GetUser)
			r.Put("/admin/users/{id}/roles", h.UpdateUserRoles)
			r.Post("/admin/revoke", h.AdminRevokeToken)
		})
	})

	return r
}
