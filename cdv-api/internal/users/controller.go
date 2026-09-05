package users

import (
	"encoding/json"
	"net/http"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

func (c *Controller) GetUsers(w http.ResponseWriter, r *http.Request) {
	usuarios, err := c.service.GetUsers(
		r.Context(),
	)

	if err != nil {

		http.Error(
			w,
			"Error al obtener usuarios",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(usuarios)
}
