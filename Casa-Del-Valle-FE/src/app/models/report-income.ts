export interface IncomeReportFilter {
  startDate: string;
  endDate: string;
  cabinId?: number;
}

export type IncomeMovementType = 'anticipo' | 'pago' | 'reembolso';

export interface IncomeMovement {
  id: number;
  reservation_id: number;
  cabin_id: number;
  cabin_name: string;
  guest_name: string;
  type: IncomeMovementType;
  method: string;
  amount: number;
  date: string;
}

export interface IncomeReportResponse {
  movements: IncomeMovement[];
  total_payments: number;
  total_refunds: number;
  total_income: number;
  total_movements: number;
  start_date: string;
  end_date: string;
  message?: string;
}
