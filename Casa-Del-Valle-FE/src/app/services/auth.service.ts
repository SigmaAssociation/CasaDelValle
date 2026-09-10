import { HttpClient } from '@angular/common/http';
import { Injectable, computed, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';
import { RestConstants } from '../components/rest-constants';
import { AuthUser, LoginRequest, LoginResponse } from '../models/auth';

@Injectable({
    providedIn: 'root',
})
export class AuthService {
    private restConstants = new RestConstants();

    private readonly tokenKey = 'cdv_token';

    private readonly tokenState = signal<string | null>(localStorage.getItem(this.tokenKey));

    readonly token = computed(() => this.tokenState());
    readonly user = computed(() => {
        const token = this.tokenState();
        return token ? this.decodeToken(token) : null;
    });
    readonly isAuthenticated = computed(() => this.tokenState() !== null);

    constructor(private httpClient: HttpClient) {}

    public login(email: string, password: string): Observable<LoginResponse> {
        const body: LoginRequest = { email, password };
        return this.httpClient
            .post<LoginResponse>(`${this.restConstants.getApiURL()}login`, body)
            .pipe(tap((response) => this.persistSession(response)));
    }

    public getToken(): string | null {
        return this.tokenState();
    }

    public getCurrentUserId(): number | null {
        const id = Number(this.user()?.id);
        return Number.isInteger(id) && id > 0 ? id : null;
    }

    public logout(): void {
        localStorage.removeItem(this.tokenKey);
        this.tokenState.set(null);
    }

    private persistSession(response: LoginResponse): void {
        if (!response.token) {
            return;
        }

        localStorage.setItem(this.tokenKey, response.token);
        this.tokenState.set(response.token);
    }

    private decodeToken(token: string): AuthUser | null {
        try {
            const payload = token.split('.')[1];
            const base64 = payload.replace(/-/g, '+').replace(/_/g, '/');
            const json = JSON.parse(decodeURIComponent(escape(atob(base64))));
            return {
                id: json['sub'],
                name: json['name'],
                email: json['email'],
                role: json['role'],
            };
        } catch {
            return null;
        }
    }
}
