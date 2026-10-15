export interface Payment {
  id: number;
  reservationId: number;
  userId: number;
  amount: number;
  method: string;
  date: string;
}
