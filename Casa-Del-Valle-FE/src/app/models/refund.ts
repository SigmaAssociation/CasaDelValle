import { Reservation } from './reservation';

export type MovementType = 'pago' | 'reembolso';

export interface FinancialMovement {
  id: number;
  type: MovementType;
  subtype: string;
  amount: number;
  method: string;
  createdAt: string;
}

export interface RefundsHistoryResponse {
  movements: FinancialMovement[];
  totalPaid: number;
  totalRefunded: number;
  refundableAvailable: number;
}

export interface Refund {
  id: number;
  reservationId: number;
  hostId: number;
  amount: number;
  method: string;
  createdAt: string;
}

export interface CreateRefundRequest {
  amount?: number;
}

export interface CreateRefundResponse {
  message: string;
  refund: Refund;
}

export interface RefundEvent {
  message: string;
  reservationId: number;
  history: RefundsHistoryResponse;
}

export type RefundTier = 'total' | 'parcial' | 'ninguno';

export const REFUND_PERCENT: Record<RefundTier, number> = {
  total: 1,
  parcial: 0.7,
  ninguno: 0,
};

export function anticipoFor(totalPrice: number): number {
  return Math.round(totalPrice * 0.3 * 100) / 100;
}

export function daysUntilStartDate(startDate: string, now: Date = new Date()): number {
  const today = new Date(now);
  today.setHours(0, 0, 0, 0);
  const start = new Date(startDate.length === 10 ? `${startDate}T00:00:00` : startDate);
  return Math.round((start.getTime() - today.getTime()) / 86_400_000);
}

export function refundTierFor(startDate: string, now: Date = new Date()): RefundTier {
  const days = daysUntilStartDate(startDate, now);
  if (days >= 7) return 'total';
  if (days >= 3) return 'parcial';
  return 'ninguno';
}

export function refundPercentFor(startDate: string, now: Date = new Date()): number {
  return REFUND_PERCENT[refundTierFor(startDate, now)];
}

export function estimatedRefundAmount(reservation: Reservation, now: Date = new Date()): number {
  if (reservation.status !== 'confirmada') return 0;
  const percent = refundPercentFor(reservation.startDate, now);
  return Math.round(anticipoFor(reservation.totalPrice) * percent * 100) / 100;
}
