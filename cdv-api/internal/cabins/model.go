package cabins

type Cabin struct {
	ID           int     `json:"id"`
	Address      string  `json:"address"`
	Price        float64 `json:"price"`
	Description  string  `json:"description"`
	Capacity     int     `json:"capacity"`
	Rules        string  `json:"rules"`
	HostID       int     `json:"host_id"`
	CommissionID *int    `json:"commission_id,omitempty"`
}

type GetCabinByIDRequest struct {
	ID int `json:"id"`
}

type GetCabinsByUserRequest struct {
	UserID int `json:"user_id"`
}

type CabinResponse struct {
	ID           int     `json:"id"`
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

type ErrorResponse struct {
	Message string `json:"message"`
}

func (c Cabin) ToResponse() CabinResponse {
	return CabinResponse{
		ID:           c.ID,
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
