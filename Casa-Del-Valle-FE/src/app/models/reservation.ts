export interface Reservation {
  id: number;
  cabinId: number;
  cabinName: string;
  cabinImageUrl?: string;
  startDate: string;
  endDate: string;
  totalPrice: number;
  status: string;
  createdAt: string;
  cancelledAt?: string | null;
}
