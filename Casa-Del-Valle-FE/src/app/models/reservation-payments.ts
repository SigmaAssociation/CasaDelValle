import { Payment } from './payment';

export interface ReservationPayments {
  reservationId: number;
  total: number;
  paid: number;
  balance: number;
  payments: Payment[];
}
