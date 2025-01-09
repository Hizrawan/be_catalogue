package controller

import (
	"net/http"
	"strconv"

	"be20250107/internal/app"
	controllers "be20250107/internal/controllers"

	"be20250107/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type CatalogueController struct {
	controllers.Controller
}

func NewCatalogueController(app *app.Registry) *CatalogueController {
	return &CatalogueController{controllers.Controller{App: app}}
}

func (c *CatalogueController) GetCatalogues(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	sortBy := r.URL.Query().Get("sortBy")
	order := r.URL.Query().Get("order")
	filterBy := r.URL.Query().Get("filterBy")
	filterValue := r.URL.Query().Get("filterValue")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10 // Default limit
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0 // Default offset
	}

	Catalogues, totalCount, err := models.GetCatalogues(c.App.DB, limit, offset, sortBy, order, filterBy, filterValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"Catalogues":  Catalogues,
		"total_count": totalCount,
	}

	render.JSON(w, r, response)
}

// GetCatalogue retrieves a single Catalogue record by ID
func (c *CatalogueController) GetCatalogue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "CatalogueID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	Catalogue, err := models.GetCatalogue(c.App.DB, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, Catalogue)
}

// CreateCatalogue creates a new Catalogue record and inserts installment values
func (c *CatalogueController) CreateCatalogue(w http.ResponseWriter, r *http.Request) {
	var Catalogue models.Catalogue
	if err := render.Bind(r, &Catalogue); err != nil {
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	tx := c.App.DB.MustBegin()
	err := Catalogue.Insert(tx)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create Catalogue record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, Catalogue)
}

// UpdateCatalogue updates an existing Catalogue record by ID and records price changes
func (c *CatalogueController) UpdateCatalogue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "CatalogueID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var Catalogue models.Catalogue
	if err := render.Bind(r, &Catalogue); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	Catalogue.ID = id

	tx := c.App.DB.MustBegin()
	err = Catalogue.Update(tx)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update Catalogue record", http.StatusInternalServerError)
		return
	}

	installment := models.CalculateInstallments(Catalogue.Price)
	installment.CatalogueID = Catalogue.ID
	// Delete existing installment records
	_, err = tx.Exec("DELETE FROM installments WHERE Catalogue_id = ?", Catalogue.ID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete existing installment records", http.StatusInternalServerError)
		return
	}

	err = installment.Insert(tx)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create installment record", http.StatusInternalServerError)
		return
	}
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, Catalogue)
}

// DeleteCatalogue deletes a Catalogue record by ID
func (c *CatalogueController) DeleteCatalogue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "CatalogueID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	Catalogue := models.Catalogue{ID: id}

	tx := c.App.DB.MustBegin()
	err = Catalogue.Delete(tx)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete Catalogue record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
