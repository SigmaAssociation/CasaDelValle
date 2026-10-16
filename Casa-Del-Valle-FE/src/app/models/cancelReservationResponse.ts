import { Reservation } from './reservation';

export interface CancelRefundInfo {
  amount: number;
  percent: number;
  kind: 'total' | 'parcial';
}

export interface CancelReservationResponse {
  message: string;
  reservation: Reservation;
  refund?: CancelRefundInfo;
}
