export interface LoginRequest {
    email: string;
    password: string;
}

export interface AuthUser {
    id: number;
    name: string;
    email: string;
    role?: number;
}

export interface LoginResponse {
    message: string;
    token: string;
}