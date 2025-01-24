package controller

import (
	"be20250107/internal/app"
	controllers "be20250107/internal/controllers"
	"be20250107/internal/responses"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

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

	hasNext := false
	if totalCount > offset+limit {
		hasNext = true
	}

	newOffset := offset + limit

	if err = responses.JSON(w, 200, struct {
		Data       []models.Catalogue           `json:"data"`
		Pagination controllers.PaginationDetail `json:"pagination"`
	}{
		Data: Catalogues,
		Pagination: controllers.PaginationDetail{
			NextPageCursor: strconv.Itoa(newOffset),
			PerPage:        limit,
			Asc:            true, // Adjust if you want to support ascending/descending
			HasNext:        hasNext,
		},
	}); err != nil {
		panic(err)
	}
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
	// Log request headers
	// log.Printf("Request Headers: %v", r.Header)

	var Catalogue models.Catalogue

	// Check if the content type is multipart form data
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		// Parse multipart form data
		if err := r.ParseMultipartForm(10 << 20); err != nil {

			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Extract form values and populate the Catalogue struct
		Catalogue.Name = r.FormValue("name")
		Catalogue.BrandID, _ = strconv.Atoi(r.FormValue("brand_id"))
		Catalogue.CategoryID, _ = strconv.Atoi(r.FormValue("category_id"))
		// Unmarshal specifications JSON string to map
		specs := r.FormValue("specifications")
		if err := json.Unmarshal([]byte(specs), &Catalogue.Specifications); err != nil {
			http.Error(w, "Invalid specifications format", http.StatusBadRequest)
			log.Printf("Error unmarshaling specifications: %v", err)
			return
		}
		// json.Unmarshal([]byte(r.FormValue("specifications")), &Catalogue.Specifications)
		Catalogue.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
		Catalogue.CreatedBy = r.FormValue("created_by")
		Catalogue.UpdatedBy = r.FormValue("updated_by")

	} else if r.Header.Get("Content-Type") == "application/json" {
		// Handle JSON payload directly
		if err := render.Bind(r, &Catalogue); err != nil {
			log.Printf("Error binding request: %v", err)
			http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Unsupported content type
		http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	// Start a new transaction
	tx, err := c.App.DB.Beginx()
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back: %v", err)
		} else {
			err = tx.Commit()
			if err != nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				log.Printf("Failed to commit transaction: %v", err)
				return
			}
		}
	}()

	// Insert catalogue with image upload handling
	err = Catalogue.Insert(tx, r)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Printf("Failed to create Catalogue record: %v", err)
		return
	}
	// Send the updated response
	response := struct {
		Ok      bool        `json:"ok"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{
		Ok:      true,
		Message: "Catalogue created successfully",
		Data:    Catalogue,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
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
