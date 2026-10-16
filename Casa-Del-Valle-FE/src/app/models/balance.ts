export type PaymentStatus = 'pendiente' | 'pagado' | 'saldo_a_favor';

export interface BalanceExtraCharge {
  id: number;
  concept: string;
  amount: number;
  createdAt: string;
}

export interface BalancePayment {
  id: number;
  amount: number;
  type: string;
  paidAt: string | null;
}

export interface BalanceRefund {
  id: number;
  paymentId: number | null;
  amount: number;
  reason: string | null;
  refundedAt: string;
}

export interface ReservationBalance {
  reservationId: number;
  status: string;
  nights: number;
  nightlyRate: number;
  baseTotal: number;
  extraChargesTotal: number;
  grandTotal: number;
  paymentsTotal: number;
  refundsTotal: number;
  netPaid: number;
  pendingBalance: number;
  paymentStatus: PaymentStatus;
  extraCharges: BalanceExtraCharge[];
  payments: BalancePayment[];
  refunds: BalanceRefund[];
}
