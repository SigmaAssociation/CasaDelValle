export interface CabinRequest {
  name: string;
  address: string;
  price: number;
  description: string;
  capacity: number;
  rules: string;
  host_id: number;
  commission_id?: number | null;
}
