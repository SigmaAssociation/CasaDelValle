export interface CabinImage {
  id: number;
  ruta: string;
  id_usuario: number;
  id_cabana: number;
}

export interface CreateImageResponse {
  mensaje: string;
  image_id: number;
  ruta: string;
}

export interface ImagesListResponse {
  data: CabinImage[];
  total: number;
}

export interface DeleteImageResponse {
  mensaje: string;
  rows_affected: number;
}