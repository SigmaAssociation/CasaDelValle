import { Cabin } from "./cabin";

// Respuesta del backend para GET /cdv-api/cabins/{id}.
// Equivale al struct CabinResponse del BE.
export interface CabinResponse extends Cabin {}

// Respuesta del backend para listados:
// GET /cdv-api/cabins y GET /cdv-api/cabins/user/{userId}.
// Equivale al struct CabinsListResponse del BE.
export interface CabinsListResponse {
    data: Cabin[];
    total: number;
}

export interface CabinCardResponse {
    id: number;
    name: string;
    address: string;
    price: number;
    capacity: number;
    host_name: string;
    host_id: number;
    image_url?: string;
}

export interface CabinCardsListResponse {
    data: CabinCardResponse[];
    total: number;
}

export interface CabinSearchParams {
    name?: string;
    host_name?: string | null;
    host_id?: number | null;
    min_capacity?: number | null;
    max_capacity?: number | null;
    min_price?: number | null;
    max_price?: number | null;
}
