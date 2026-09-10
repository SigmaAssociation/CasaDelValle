import { HttpClient } from "@angular/common/http";
import { Injectable, signal } from "@angular/core";
import { Router } from "@angular/router";
import { Observable, tap } from "rxjs";
import { RestConstants } from "../components/rest-constants";
import { LoginRequest, LoginResponse } from "../models/login";

const TOKEN_KEY = 'cdv_token';

interface JwtPayload {
    sub?: number | string;
    name?: string;
    email?: string;
    role?: number;
    exp?: number;
}

@Injectable({
    providedIn: 'root',
})
export class AuthService {
    restConstants = new RestConstants();

    isLoggedIn = signal(this.hasValidToken());

    constructor(
        private httpClient: HttpClient,
        private router: Router
    ) { }

    public login(payload: LoginRequest): Observable<LoginResponse> {
        return this.httpClient.post<LoginResponse>(
            `${this.restConstants.getApiURL()}login`, payload
        ).pipe(
            tap((res) => {
                if (res?.token) {
                    this.saveToken(res.token);
                }
            })
        );
    }

    public logout(): void {
        localStorage.removeItem(TOKEN_KEY);
        this.isLoggedIn.set(false);
        this.router.navigate(['/']);
    }

    public getToken(): string | null {
        return localStorage.getItem(TOKEN_KEY);
    }

    public getCurrentUserId(): number | null {
        const payload = this.decodePayload();
        if (!payload?.sub) return null;
        const id = Number(payload.sub);
        return Number.isInteger(id) && id > 0 ? id : null;
    }

    public getCurrentUserName(): string | null {
        return this.decodePayload()?.name ?? null;
    }

    private saveToken(token: string): void {
        localStorage.setItem(TOKEN_KEY, token);
        this.isLoggedIn.set(true);
    }

    private hasValidToken(): boolean {
        const payload = this.decodePayload();
        if (!payload) return false;
        if (payload.exp && payload.exp * 1000 < Date.now()) return false;
        return true;
    }

    private decodePayload(): JwtPayload | null {
        try {
            const token = localStorage.getItem(TOKEN_KEY);
            if (!token) return null;
            const parts = token.split('.');
            if (parts.length !== 3) return null;
            const json = atob(parts[1].replace(/-/g, '+').replace(/_/g, '/'));
            return JSON.parse(json) as JwtPayload;
        } catch {
            return null;
        }
    }
}
