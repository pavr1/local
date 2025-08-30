package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"inventory-service/entities/recipe_ingredients/models"
	httpresponse "shared/http-response"
	sharedLogger "shared/logger"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type HttpHandler struct {
	dbHandler DBHandlerInterface
}

// DBHandlerInterface defines the database operations interface
type DBHandlerInterface interface {
	Create(req models.CreateRecipeIngredientRequest, logger *logrus.Logger) (*models.RecipeIngredient, error)
	GetByID(req models.GetRecipeIngredientRequest, logger *logrus.Logger) (*models.RecipeIngredient, error)
	List(req models.ListRecipeIngredientsRequest, logger *logrus.Logger) ([]models.RecipeIngredient, error)
	Update(req models.UpdateRecipeIngredientRequest, id string, logger *logrus.Logger) (*models.RecipeIngredient, error)
	Delete(req models.DeleteRecipeIngredientRequest, logger *logrus.Logger) error
}

// Ensure DBHandler implements DBHandlerInterface
var _ DBHandlerInterface = (*DBHandler)(nil)

// NewHttpHandler creates a new HTTP handler
func NewHttpHandler(dbHandler DBHandlerInterface) *HttpHandler {
	return &HttpHandler{
		dbHandler: dbHandler,
	}
}

// CreateRecipeIngredient handles POST /recipe-ingredients
func (h *HttpHandler) CreateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	var req models.CreateRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in create recipe ingredient request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	recipeIngredient, err := h.dbHandler.Create(req, logger.Logger)
	if err != nil {
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.RecipeIngredient{},
			Message: "Failed to create recipe ingredient: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusCreated,
		Data:    *recipeIngredient,
		Message: "Recipe ingredient created successfully",
	}

	logger.WithFields(logrus.Fields{
		"recipe_ingredient_id": recipeIngredient.ID,
		"recipe_id":            recipeIngredient.RecipeID,
		"ingredient_id":        recipeIngredient.IngredientID,
	}).Info("Recipe ingredient created successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// GetRecipeIngredient handles GET /recipe-ingredients/{id}
func (h *HttpHandler) GetRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing recipe ingredient ID in get request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe ingredient ID is required",
		})
		return
	}

	req := models.GetRecipeIngredientRequest{ID: id}
	recipeIngredient, err := h.dbHandler.GetByID(req, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.RecipeIngredient{},
				Message: "Recipe ingredient not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.RecipeIngredient{},
			Message: "Failed to get recipe ingredient: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *recipeIngredient,
		Message: "Recipe ingredient retrieved successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// ListRecipeIngredients handles GET /recipe-ingredients
func (h *HttpHandler) ListRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	req := models.ListRecipeIngredientsRequest{}

	// Parse query parameters
	if recipeID := r.URL.Query().Get("recipe_id"); recipeID != "" {
		req.RecipeID = &recipeID
	}

	if ingredientID := r.URL.Query().Get("ingredient_id"); ingredientID != "" {
		req.IngredientID = &ingredientID
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = &limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = &offset
		}
	}

	recipeIngredients, err := h.dbHandler.List(req, logger.Logger)
	if err != nil {
		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    []models.RecipeIngredient{},
			Message: "Failed to list recipe ingredients: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    recipeIngredients,
		Message: "Recipe ingredients retrieved successfully",
	}
	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// UpdateRecipeIngredient handles PUT /recipe-ingredients/{id}
func (h *HttpHandler) UpdateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing recipe ingredient ID in update request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe ingredient ID is required",
		})
		return
	}

	var req models.UpdateRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("Invalid JSON in update recipe ingredient request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}

	recipeIngredient, err := h.dbHandler.Update(req, id, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"recipe_ingredient_id": id,
			}).Warn("Recipe ingredient not found for update")
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Data:    models.RecipeIngredient{},
				Message: "Recipe ingredient not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Data:    models.RecipeIngredient{},
			Message: "Failed to update recipe ingredient: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Data:    *recipeIngredient,
		Message: "Recipe ingredient updated successfully",
	}

	logger.WithFields(logrus.Fields{
		"recipe_ingredient_id": recipeIngredient.ID,
		"recipe_id":            recipeIngredient.RecipeID,
		"ingredient_id":        recipeIngredient.IngredientID,
	}).Info("Recipe ingredient updated successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}

// DeleteRecipeIngredient handles DELETE /recipe-ingredients/{id}
func (h *HttpHandler) DeleteRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	// Get logger with request ID
	logger := sharedLogger.GetRequestLogger(r, sharedLogger.SERVICE_INVENTORY_SERVICE)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		logger.Warn("Missing recipe ingredient ID in delete request")
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, httpresponse.Response{
			Code:    http.StatusBadRequest,
			Message: "Recipe ingredient ID is required",
		})
		return
	}

	req := models.DeleteRecipeIngredientRequest{ID: id}
	err := h.dbHandler.Delete(req, logger.Logger)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"recipe_ingredient_id": id,
			}).Warn("No recipe ingredient found to delete")
			response := httpresponse.Response{
				Code:    http.StatusNotFound,
				Message: "Recipe ingredient not found",
			}
			httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
			return
		}

		response := httpresponse.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete recipe ingredient: " + err.Error(),
		}
		httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
		return
	}

	response := httpresponse.Response{
		Code:    http.StatusOK,
		Message: "Recipe ingredient deleted successfully",
	}

	logger.WithFields(logrus.Fields{
		"recipe_ingredient_id": id,
	}).Info("Recipe ingredient deleted successfully")

	httpresponse.WriteResponse(w, r, sharedLogger.SERVICE_INVENTORY_SERVICE, response)
}
