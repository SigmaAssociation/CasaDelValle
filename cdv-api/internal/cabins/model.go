package cabins

type Cabin struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	Price        float64 `json:"price"`
	Description  string  `json:"description"`
	Capacity     int     `json:"capacity"`
	Rules        string  `json:"rules"`
	HostID       int     `json:"host_id"`
	CommissionID *int    `json:"commission_id,omitempty"`
}

type CreateCabinRequest struct {
	Nombre      string `json:"nombre"`
	Direccion  string  `json:"direccion"`
	Precio     float64 `json:"precio"`
	Descripcion string `json:"descripcion,omitempty"`
	Capacidad  int     `json:"capacidad"`
	Reglas     string  `json:"reglas,omitempty"`
	IDAnfitrion int    `json:"id_anfitrion"`
	IDComision *int    `json:"id_comision,omitempty"`
}

type CreateCabinResponse struct {
	Message string `json:"mensaje"`
	CabinID int    `json:"cabin_id,omitempty"`
}

type GetCabinByIDRequest struct {
	ID int `json:"id"`
}

type GetCabinsByUserRequest struct {
	UserID int `json:"user_id"`
}

type CabinResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	Price        float64 `json:"price"`
	Description  string  `json:"description"`
	Capacity     int     `json:"capacity"`
	Rules        string  `json:"rules"`
	HostID       int     `json:"host_id"`
	CommissionID *int    `json:"commission_id,omitempty"`
}

type CabinsListResponse struct {
	Data  []CabinResponse `json:"data"`
	Total int             `json:"total"`
}

/*Modelo para la info completa en las card de cabañas*/
type CabinCardResponse struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Price    float64 `json:"price"`
	Capacity int     `json:"capacity"`
	HostName string  `json:"host_name"`
	ImageURL *string `json:"image_url,omitempty"`
}

type CabinCardsListResponse struct {
	Data  []CabinCardResponse `json:"data"`
	Total int                 `json:"total"`
}

type CabinSearchParams struct {
	HostID      *int     `json:"host_id,omitempty"`
	MinCapacity *int     `json:"min_capacity,omitempty"`
	MaxCapacity *int     `json:"max_capacity,omitempty"`
	MinPrice    *float64 `json:"min_price,omitempty"`
	MaxPrice    *float64 `json:"max_price,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type GetCabinsByCapacityRequest struct {
    MinCapacity int `json:"min_capacity"`
    MaxCapacity int `json:"max_capacity"`
}

type GetCabinsByPriceRangeRequest struct {
    MinPrice float64 `json:"min_price"`
    MaxPrice float64 `json:"max_price"`
}

func (c Cabin) ToResponse() CabinResponse {
	return CabinResponse{
		ID:           c.ID,
		Name:         c.Name,
		Address:      c.Address,
		Price:        c.Price,
		Description:  c.Description,
		Capacity:     c.Capacity,
		Rules:        c.Rules,
		HostID:       c.HostID,
		CommissionID: c.CommissionID,
	}
}

func ToResponseList(cabins []Cabin) CabinsListResponse {
	data := make([]CabinResponse, 0, len(cabins))
	for _, c := range cabins {
		data = append(data, c.ToResponse())
	}
	return CabinsListResponse{
		Data:  data,
		Total: len(data),
	}
}

func (c Cabin) ToCardResponse(hostName string, imageURL *string) CabinCardResponse {
	return CabinCardResponse{
		ID:       c.ID,
		Name:     c.Name,
		Address:  c.Address,
		Price:    c.Price,
		Capacity: c.Capacity,
		HostName: hostName,
		ImageURL: imageURL,
	}
}