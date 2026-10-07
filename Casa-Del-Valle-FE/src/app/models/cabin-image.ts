export interface CabinImage {
    id: number;
    path: string;
    user_id?: number | null;
    cabin_id?: number | null;
    created_at: string;
}

export interface CabinImagesListResponse {
    data: CabinImage[];
    total: number;
}