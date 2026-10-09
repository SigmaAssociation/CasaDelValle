export interface Reservation {
  id: number;
  cabinId: number;
  cabinName: string;
  cabinImageUrl?: string;
  guestId?: number;
  guestName?: string;
  startDate: string;
  endDate: string;
  totalPrice: number;
  status: string;
  createdAt: string;
  cancelledAt?: string | null;
}
