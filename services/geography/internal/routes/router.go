package routes

import (
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/geography/internal/config"

	"gorm.io/gorm"
)

func RegisterRoutes(cfg *config.Config, db *gorm.DB) *chi.Mux {
	controller := NewGeographyController(db)

	mux := chi.NewRouter()

	// Health check
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Countries
	mux.Route("/api/countries", func(r chi.Router) {
		r.Get("/", controller.ListCountries)
		r.Post("/", controller.CreateCountry)
		r.Get("/{id}", controller.GetCountry)
		r.Put("/{id}", controller.UpdateCountry)
		r.Delete("/{id}", controller.DeleteCountry)
	})

	// Cities
	mux.Route("/api/cities", func(r chi.Router) {
		r.Get("/", controller.ListCities)
		r.Post("/", controller.CreateCity)
		r.Get("/{id}", controller.GetCity)
		r.Put("/{id}", controller.UpdateCity)
		r.Delete("/{id}", controller.DeleteCity)
	})

	// Locations
	mux.Route("/api/locations", func(r chi.Router) {
		r.Get("/", controller.ListLocations)
		r.Post("/", controller.CreateLocation)
		r.Get("/{id}", controller.GetLocation)
		r.Put("/{id}", controller.UpdateLocation)
		r.Delete("/{id}", controller.DeleteLocation)
	})

	// Contacts
	mux.Route("/api/contacts", func(r chi.Router) {
		r.Get("/", controller.ListContacts)
		r.Post("/", controller.CreateContact)
		r.Get("/{id}", controller.GetContact)
		r.Put("/{id}", controller.UpdateContact)
		r.Delete("/{id}", controller.DeleteContact)
	})

	return mux
}

