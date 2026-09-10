import { HttpClient } from '@angular/common/http';
import { Injectable, computed, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';
import { RestConstants } from '../components/rest-constants';
import { AuthUser, LoginRequest, LoginResponse } from '../models/auth';

export interface AuthState {
    token: string | null;
    user: AuthUser | null;
}

@Injectable({
    providedIn: 'root',
})
export class AuthService {
    private restConstants = new RestConstants();

    private readonly tokenKey = 'cdv_token';
    private readonly userKey = 'cdv_user';

    private readonly state = signal<AuthState>(this.loadState());

    readonly token = computed(() => this.state().token);
    readonly user = computed(() => this.state().user);
    readonly isAuthenticated = computed(() => this.state().token !== null);

    constructor(private httpClient: HttpClient) {}

    public login(email: string, password: string): Observable<LoginResponse> {
        const body: LoginRequest = { email, password };
        return this.httpClient
            .post<LoginResponse>(`${this.restConstants.getApiURL()}login`, body)
            .pipe(tap((response) => this.persistSession(response)));
    }

    public getToken(): string | null {
        return this.state().token;
    }

    public logout(): void {
        localStorage.removeItem(this.tokenKey);
        localStorage.removeItem(this.userKey);
        this.state.set({ token: null, user: null });
    }

    private persistSession(response: LoginResponse): void {
        if (!response.token) {
            return;
        }

        const user = this.decodeToken(response.token);

        localStorage.setItem(this.tokenKey, response.token);
        if (user) {
            localStorage.setItem(this.userKey, JSON.stringify(user));
        }

        this.state.set({ token: response.token, user });
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

    private loadState(): AuthState {
        const token = localStorage.getItem(this.tokenKey);
        if (!token) {
            return { token: null, user: null };
        }

        const rawUser = localStorage.getItem(this.userKey);
        try {
            return {
                token,
                user: rawUser ? (JSON.parse(rawUser) as AuthUser) : this.decodeToken(token),
            };
        } catch {
            return { token, user: null };
        }
    }
}