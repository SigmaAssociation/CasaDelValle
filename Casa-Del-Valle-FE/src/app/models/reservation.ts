export interface Reservation {
  id: number;
  cabinId: number;
  hostId?: number;
  cabinName: string;
  cabinImageUrl?: string;
  guestId?: number;
  guestName?: string;
  startDate: string;
  endDate: string;
  totalPrice: number;
  paidTotal?: number;
  status: string;
  createdAt: string;
  cancelledAt?: string | null;
}
