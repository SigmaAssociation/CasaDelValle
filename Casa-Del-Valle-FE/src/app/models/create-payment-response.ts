import { Payment } from './payment';

export interface CreatePaymentResponse {
  message: string;
  payment: Payment;
  reservationId: number;
  total: number;
  paid: number;
  balance: number;
  payments: Payment[];
}
