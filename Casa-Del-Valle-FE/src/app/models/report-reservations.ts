export type EstadoReservacion =
  | 'pendiente'
  | 'confirmada'
  | 'en_curso'
  | 'finalizada'
  | 'cancelada';

export interface ReservationsReportFilter {
  host_id?: number;
  cabin_id?: number;
}

export interface UsuarioResumen {
  id: number;
  nombre: string;
}

export interface ReservacionResumen {
  id: number;
  fecha_inicio: string; 
  fecha_fin: string;    
  estado: EstadoReservacion;
  fecha_creacion: string; 
  huesped: UsuarioResumen;
}

export interface CabanaReporte {
  id: number;
  nombre: string;
  direccion: string;
  precio: number;
  capacidad: number;
  anfitrion: UsuarioResumen;
  total_reservaciones: number;
  reservaciones: ReservacionResumen[];
}

export interface ReservationsReportResponse {
  cabanas: CabanaReporte[];
  total_cabanas: number;
  total_reservaciones: number;
  mensaje?: string; 
}

export interface ApiError {
  message: string;
}