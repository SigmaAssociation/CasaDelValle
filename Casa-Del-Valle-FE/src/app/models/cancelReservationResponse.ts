import { Reservation } from "./reservation";

export interface CancelReservationResponse {
  message: string;
  reservation: Reservation;
}