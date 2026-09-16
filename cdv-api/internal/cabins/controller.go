package cabins

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cdv-api/internal/middleware"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Message: message})
}

func (c *Controller) GetCabins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	cabins, err := c.service.GetCabins(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener las cabañas")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(cabins))
}

func (c *Controller) CreateCabin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var req CreateCabinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de entrada inválidos")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(req.IDAnfitrion)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para registrar esta cabaña")
		return
	}

	cabinID, err := c.service.CreateCabin(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CreateCabinResponse{
		Message: "Cabaña registrada exitosamente",
		CabinID: cabinID,
	})
}

func (c *Controller) GetCabinByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("id")
	if idParam == "" {
		idParam = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID de cabaña inválido")
		return
	}

	cabin, err := c.service.GetCabinByID(r.Context(), GetCabinByIDRequest{ID: id})
	if err != nil {
		if err.Error() == "Cabaña no encontrada" {
			writeError(w, http.StatusNotFound, "Cabaña no encontrada")
			return
		}

		if err.Error() == "ID de cabaña inválido" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "Error al obtener la cabaña")
		return
	}

	writeJSON(w, http.StatusOK, cabin.ToResponse())
}

func (c *Controller) GetCabinsByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	idParam := r.PathValue("userId")
	if idParam == "" {
		idParam = r.URL.Query().Get("userId")
	}

	userID, err := strconv.Atoi(idParam)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "ID de usuario inválido")
		return
	}

	if !middleware.AuthorizeSelfOrAdmin(r, uint(userID)) {
		writeError(w, http.StatusForbidden, "No tiene permisos para ver las cabañas de este usuario")
		return
	}

	cabins, err := c.service.GetCabinsByUser(
		r.Context(),
		GetCabinsByUserRequest{UserID: userID},
	)
	if err != nil {
		if err.Error() == "ID de usuario inválido" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "Error al obtener las cabañas del usuario")
		return
	}

	writeJSON(w, http.StatusOK, ToResponseList(cabins))
}

func (c *Controller) GetAllCabins(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
        return
    }

    cabins, err := c.service.GetAllCabins(r.Context())
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Error al obtener las cabañas")
        return
    }

    writeJSON(w, http.StatusOK, ToResponseList(cabins))
}

func (c *Controller) GetCabinsByCapacity(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
        return
    }

    minCap, errMin := strconv.Atoi(r.URL.Query().Get("min_capacity"))
    maxCap, errMax := strconv.Atoi(r.URL.Query().Get("max_capacity"))

    if errMin != nil || errMax != nil {
        writeError(w, http.StatusBadRequest, "Los parámetros de capacidad deben ser números enteros válidos")
        return
    }

    req := GetCabinsByCapacityRequest{
        MinCapacity: minCap,
        MaxCapacity: maxCap,
    }

    cabins, err := c.service.GetCabinsByCapacity(r.Context(), req)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, ToResponseList(cabins))
}

func (c *Controller) GetCabinsByPrice(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
        return
    }

    minPrice, errMin := strconv.ParseFloat(r.URL.Query().Get("min_price"), 64)
    maxPrice, errMax := strconv.ParseFloat(r.URL.Query().Get("max_price"), 64)

    if errMin != nil || errMax != nil {
        writeError(w, http.StatusBadRequest, "Los parámetros de precio deben ser números válidos")
        return
    }

    req := GetCabinsByPriceRangeRequest{
        MinPrice: minPrice,
        MaxPrice: maxPrice,
    }

    cabins, err := c.service.GetCabinsByPrice(r.Context(), req)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, ToResponseList(cabins))
}