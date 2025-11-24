package routes

import (
	"net/http"

	"github.com/go-chi/chi"

	"backend/config"
	"backend/controllers"
	"backend/services"
)

/* func RegisterRoutesOld(cfg *config.Config) http.Handler {
	userService := services.NewUserService(cfg.DB)
	authController := controllers.NewAuthController(cfg, userService)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/login", authController.Login)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Protected routes group
	// Use JWT middleware globally for a subrouter:
	//api := mux.PathPrefix("/api").Subrouter()
	//api.Use(middleware.JWTAuthMiddleware(cfg))

	//Register Routes From files

	return mux
} */

func RegisterRoutes(cfg *config.Config) http.Handler {
	userService := services.NewUserService(cfg.DB)
	authController := controllers.NewAuthController(cfg, userService)

	mux := chi.NewRouter()

	// Public routes
	mux.HandleFunc("/login", authController.Login)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Protected routes group
	// Use JWT middleware globally for a subrouter:
	//api := mux.PathPrefix("/api").Subrouter()
	//api.Use(middleware.JWTAuthMiddleware(cfg))

	//Register Routes From files
	RegisterPermissionRoutes(mux, cfg)
	RegisterAdminRoutes(mux, cfg)
	RegisterRoleRoutes(mux, cfg)
	RegisterUserRoutes(mux, cfg)
	RegisterHelpDeskTypeRoutes(mux, cfg)
	RegisterInventoryRoutes(mux, cfg)
	RegisterGeographyRoutes(mux, cfg)
	RegisterHelpDeskRoutes(mux, cfg)
	RegisterNavigationRoutes(mux, cfg)

	return mux
}
