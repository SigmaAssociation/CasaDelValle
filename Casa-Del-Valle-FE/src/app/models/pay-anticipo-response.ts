import { Reservation } from "./reservation";

export interface PayAnticipoResponse {
  message: string;
  amount: number;
  reservation: Reservation;
}
