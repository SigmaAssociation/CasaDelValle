export interface Cabin {
    id: number;
    name: string;
    address: string;
    price: number;
    description: string;
    capacity: number;
    rules: string;
    host_id?: number;
    commission_id?: number | null;
}