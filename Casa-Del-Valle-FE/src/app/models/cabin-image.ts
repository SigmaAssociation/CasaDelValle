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

export interface CreateImageResponse {
    message: string;
    image_id?: number;
    path?: string;
}

export interface DeleteImageResponse {
    message: string;
    rows_affected?: number;
}
