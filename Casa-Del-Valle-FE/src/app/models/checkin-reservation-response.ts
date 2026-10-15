import { ReservationDetail } from "./reservation-detail";

export interface CheckinReservationResponse {
  message: string;
  reservation: ReservationDetail;
}
