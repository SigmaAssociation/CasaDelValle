package images

import "time"

type Image struct {
	ID        int
	Path      string
	UserID    *int
	CabinID   *int
	CreatedAt time.Time
}

// DTO de salida para imagen.
type ImageResponse struct {
	ID        int       `json:"id"`
	Path      string    `json:"path"`
	UserID    *int      `json:"user_id,omitempty"`
	CabinID   *int      `json:"cabin_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// DTO de salida para listados de imágenes.
type ImagesListResponse struct {
	Data  []ImageResponse `json:"data"`
	Total int             `json:"total"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func (i Image) ToResponse() ImageResponse {
	return ImageResponse{
		ID:        i.ID,
		Path:      i.Path,
		UserID:    i.UserID,
		CabinID:   i.CabinID,
		CreatedAt: i.CreatedAt,
	}
}

func ToResponseList(images []Image) ImagesListResponse {
	data := make([]ImageResponse, 0, len(images))
	for _, i := range images {
		data = append(data, i.ToResponse())
	}
	return ImagesListResponse{
		Data:  data,
		Total: len(data),
	}
}
