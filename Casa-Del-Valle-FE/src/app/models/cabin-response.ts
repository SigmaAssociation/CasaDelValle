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
