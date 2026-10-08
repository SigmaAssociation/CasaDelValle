export interface ReservationRequest {
  user_id: number;
  cabin_id: number;
  start_date: string; // Formato AAAA-MM-DD
  end_date: string;   // Formato AAAA-MM-DD
}
