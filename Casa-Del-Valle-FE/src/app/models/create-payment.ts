import { PaymentMethod } from './payment-method';

export interface CreatePayment {
  amount: number;
  method: PaymentMethod;
}
