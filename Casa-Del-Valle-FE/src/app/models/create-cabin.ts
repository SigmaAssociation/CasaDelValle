export interface CabinRequest {
  direccion: string;
  precio: number;
  descripcion: string;
  capacidad: number;
  reglas: string;
  id_anfitrion: number;
  id_comision?: number | null;
}
