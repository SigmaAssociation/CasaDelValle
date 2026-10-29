export interface Maintenance {
  id: number; cabin_id: number; start_date: string; end_date: string;
  reason: string; created_at: string; cancelled_at: string | null;
}
export interface AffectedReservation {
  id: number; guest_id: number; guest_name: string; start_date: string; end_date: string;
  status: string; total_price: number;
}
export interface CreateMaintenance { start_date: string; end_date: string; reason: string; }
export interface MaintenanceDecision {
  action: 'reprogramar' | 'reembolsar'; start_date?: string; end_date?: string;
}
export interface PendingMaintenance {
  reservation_id: number; cabin_id: number; reason: string; start_date: string; end_date: string;
  nights: number; total_price: number; refundable: number;
  blocked_ranges: { start: string; end: string }[];
}
