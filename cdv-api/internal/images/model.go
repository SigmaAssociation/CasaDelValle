package images

type Image struct {
	ID       int    `json:"id"`
	Ruta     string `json:"ruta"`
	IDUsuario int   `json:"id_usuario,omitempty"`
	IDCabana int    `json:"id_cabana"`
}

type CreateImageRequest struct {
	IDCabana int `json:"id_cabana"`
}

type CreateImageResponse struct {
	Message string `json:"mensaje"`
	ImageID int    `json:"image_id,omitempty"`
	Ruta    string `json:"ruta"`
}

type ImagesListResponse struct {
	Data  []Image `json:"data"`
	Total int     `json:"total"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}