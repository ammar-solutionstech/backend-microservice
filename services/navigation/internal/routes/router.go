package routes

import (
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/navigation/internal/config"
	"gorm.io/gorm"
)

func RegisterRoutes(cfg *config.Config, db *gorm.DB) *chi.Mux {
	controller := NewNavigationController(db)

	mux := chi.NewRouter()

	// Health check
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Menus
	mux.Route("/api/menus", func(r chi.Router) {
		r.Get("/", controller.ListMenus)
		r.Post("/", controller.CreateMenu)
		r.Get("/{id}", controller.GetMenu)
		r.Put("/{id}", controller.UpdateMenu)
		r.Delete("/{id}", controller.DeleteMenu)

		// Menu Roles
		r.Route("/{menuId}/roles", func(sr chi.Router) {
			sr.Get("/", controller.GetMenuRoles)
			sr.Post("/", controller.AddMenuRole)
			sr.Delete("/{roleId}", controller.RemoveMenuRole)
		})
	})

	return mux
}

