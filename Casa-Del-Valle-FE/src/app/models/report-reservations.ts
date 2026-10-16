export type ReservationStatus =
  | 'pendiente'
  | 'confirmada'
  | 'en_curso'
  | 'finalizada'
  | 'cancelada';

export interface ReservationsReportFilter {
  host_id?: number;
  cabin_id?: number;
  guest_id?: number;
  status?: ReservationStatus;
  created_from?: string;
  created_to?: string;
}

export interface UserSummary {
  id: number;
  name: string;
}

export interface ReservationSummary {
  id: number;
  start_date: string;
  end_date: string;
  status: ReservationStatus;
  created_at: string;
  guest: UserSummary;
}

export interface CabinReport {
  id: number;
  name: string;
  address: string;
  price: number;
  capacity: number;
  host: UserSummary;
  total_reservations: number;
  reservations: ReservationSummary[];
}

export interface ReservationsReportResponse {
  cabins: CabinReport[];
  total_cabins: number;
  total_reservations: number;
  message?: string;
}

export interface ApiError {
  message: string;
}