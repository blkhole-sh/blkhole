package routes

import (
	"github.com/lemon3studio/blkhole/internal/controllers"
	"github.com/lemon3studio/blkhole/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

// Init initializes the api routes
func InitAPI(
	r *chi.Mux,
	authController controllers.AuthController,
	userController controllers.UserController,
	deviceController controllers.DeviceController,
	listController controllers.ListController,
	scheduleController controllers.ScheduleController,
	statsController controllers.StatsController,
	quoteController controllers.QuoteController,
	mobileConfigController controllers.MobileConfigController,
	settingsController controllers.SettingsController,
	testController controllers.TestController,
	webController controllers.WebController,
	queryLogController controllers.QueryLogController,
	tokenAuth *jwtauth.JWTAuth,
) {
	// API routes group
	r.Route("/api", func(r chi.Router) {
		// Apply JSON middlew to all API routes
		r.Use(middleware.JSONMiddleware)

		// Public routes
		r.Post("/auth/register", authController.Register)
		r.Post("/auth/login", authController.Login)
		r.Post("/auth/refresh", authController.RefreshToken)
		r.Post("/auth/logout", authController.Logout)
		r.Get("/quote", quoteController.Random)
		r.Get("/settings", settingsController.GetSettings)

		// Protected routes
		r.Group(func(r chi.Router) {
			// Cookie-based JWT authentication middlew
			r.Use(middleware.CookieAuthenticator(tokenAuth))

			// Auth routes
			r.Get("/auth/me", authController.GetCurrentUser)

			// User API routes
			r.Get("/users/{id}", userController.FindByID)
			r.Put("/users", userController.Create)
			r.Patch("/users/{id}", userController.Update)
			r.Delete("/users/{id}", userController.Delete)

			// Device API routes
			r.Get("/devices/{id}", deviceController.FindByID)
			r.Get("/devices/{id}/config", mobileConfigController.GenerateConfig)
			r.Get("/users/{userId}/devices", deviceController.FindByUser)
			r.Put("/devices", deviceController.Create)
			r.Patch("/devices/{id}", deviceController.Update)
			r.Delete("/devices/{id}", deviceController.Delete)

			// List API routes
			r.Get("/lists/{id}", listController.FindByID)
			r.Get("/users/{userId}/lists", listController.FindByUser)
			r.Put("/lists", listController.Create)
			r.Patch("/lists/{id}", listController.Update)
			r.Delete("/lists/{id}", listController.Delete)

			// Schedule API routes
			r.Get("/schedules/{id}", scheduleController.FindByID)
			r.Get("/users/{userId}/schedules", scheduleController.FindByUser)
			r.Put("/schedules", scheduleController.Create)
			r.Patch("/schedules/{id}", scheduleController.Update)
			r.Delete("/schedules/{id}", scheduleController.Delete)
			r.Get("/is-blocked", scheduleController.IsBlocked)

			// Stats API routes
			r.Get("/users/{userId}/stats/queries", statsController.GetQueryStats)

			// Query log routes
			r.Get("/users/{userId}/logs", queryLogController.GetLogs)
			r.Get("/users/{userId}/logs/export", queryLogController.ExportLogs)
		})
	})

	// Serve test route
	r.Get("/test", testController.RunTest)
}
