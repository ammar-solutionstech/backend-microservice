package routes

import (
	"net/http"

	"github.com/go-chi/chi"

	"backend/services/inventory/internal/config"

	"gorm.io/gorm"
)

func RegisterRoutes(cfg *config.Config, db *gorm.DB) *chi.Mux {
	controller := NewInventoryController(db)

	mux := chi.NewRouter()

	// Health check
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Brands
	mux.Route("/api/brands", func(r chi.Router) {
		r.Get("/", controller.ListBrands)
		r.Post("/", controller.CreateBrand)
		r.Get("/{id}", controller.GetBrand)
		r.Put("/{id}", controller.UpdateBrand)
		r.Delete("/{id}", controller.DeleteBrand)
	})

	// Models - using generic handlers
	mux.Route("/api/models", func(r chi.Router) {
		r.Get("/", controller.ListModels)
		r.Post("/", controller.CreateModel)
		r.Get("/{id}", controller.GetModel)
		r.Put("/{id}", controller.UpdateModel)
		r.Delete("/{id}", controller.DeleteModel)
	})

	// Equipment Types
	mux.Route("/api/equipment-types", func(r chi.Router) {
		r.Get("/", controller.ListEquipmentTypes)
		r.Post("/", controller.CreateEquipmentType)
		r.Get("/{id}", controller.GetEquipmentType)
		r.Put("/{id}", controller.UpdateEquipmentType)
		r.Delete("/{id}", controller.DeleteEquipmentType)
	})

	// Operating Systems
	mux.Route("/api/operating-systems", func(r chi.Router) {
		r.Get("/", controller.ListOperatingSystems)
		r.Post("/", controller.CreateOperatingSystem)
		r.Get("/{id}", controller.GetOperatingSystem)
		r.Put("/{id}", controller.UpdateOperatingSystem)
		r.Delete("/{id}", controller.DeleteOperatingSystem)
	})

	// Software Categories
	mux.Route("/api/software-categories", func(r chi.Router) {
		r.Get("/", controller.ListSoftwareCategories)
		r.Post("/", controller.CreateSoftwareCategory)
		r.Get("/{id}", controller.GetSoftwareCategory)
		r.Put("/{id}", controller.UpdateSoftwareCategory)
		r.Delete("/{id}", controller.DeleteSoftwareCategory)
	})

	// Software
	mux.Route("/api/software", func(r chi.Router) {
		r.Get("/", controller.ListSoftware)
		r.Post("/", controller.CreateSoftware)
		r.Get("/{id}", controller.GetSoftware)
		r.Put("/{id}", controller.UpdateSoftware)
		r.Delete("/{id}", controller.DeleteSoftware)
	})

	// Equipment
	mux.Route("/api/equipment", func(r chi.Router) {
		r.Get("/", controller.ListEquipment)
		r.Post("/", controller.CreateEquipment)
		r.Get("/{id}", controller.GetEquipment)
		r.Put("/{id}", controller.UpdateEquipment)
		r.Delete("/{id}", controller.DeleteEquipment)

		// Equipment Software relationships
		r.Route("/{equipmentId}/software", func(sr chi.Router) {
			sr.Get("/", controller.GetEquipmentSoftware)
			sr.Post("/", controller.CreateEquipmentSoftware)
			sr.Put("/{softwareId}", controller.UpdateEquipmentSoftware)
			sr.Delete("/{softwareId}", controller.DeleteEquipmentSoftware)
		})

		// Equipment Help Desk relationships
		r.Route("/{equipmentId}/help-desk", func(sr chi.Router) {
			sr.Get("/", controller.GetEquipmentHelpDeskLinks)
			sr.Post("/", controller.CreateEquipmentHelpDeskLink)
			sr.Delete("/{helpDeskId}", controller.DeleteEquipmentHelpDeskLink)
		})

		// Equipment User History
		r.Route("/{equipmentId}/user-history", func(sr chi.Router) {
			sr.Get("/", controller.GetEquipmentUserHistory)
			sr.Post("/", controller.CreateEquipmentUserHistory)
			sr.Put("/{userId}/{startDate}", controller.UpdateEquipmentUserHistory)
			sr.Delete("/{userId}/{startDate}", controller.DeleteEquipmentUserHistory)
		})
	})

	// Documents
	mux.Route("/api/documents", func(r chi.Router) {
		r.Get("/", controller.ListDocuments)
		r.Post("/", controller.CreateDocument)
		r.Get("/{id}", controller.GetDocument)
		r.Put("/{id}", controller.UpdateDocument)
		r.Delete("/{id}", controller.DeleteDocument)
	})

	// Maintenance
	mux.Route("/api/maintenance", func(r chi.Router) {
		r.Get("/", controller.ListMaintenance)
		r.Post("/", controller.CreateMaintenance)
		r.Get("/{id}", controller.GetMaintenance)
		r.Put("/{id}", controller.UpdateMaintenance)
		r.Delete("/{id}", controller.DeleteMaintenance)
	})

	return mux
}

