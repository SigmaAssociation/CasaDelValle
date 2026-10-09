export interface ReservationDetail {
  id: number;
  user_id: number;
  cabin_id: number;
  host_id: number;
  start_date: string;
  end_date: string;
  status: string;
  cancelled_at?: string | null;
}
