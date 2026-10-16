import { ReservationDetail } from './reservation-detail';

export interface CheckoutReservationResponse {
  message: string;
  reservation: ReservationDetail;
}