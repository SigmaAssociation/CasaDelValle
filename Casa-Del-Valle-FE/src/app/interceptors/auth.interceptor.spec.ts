import { TestBed } from '@angular/core/testing';
import { HttpClient, HttpErrorResponse, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { Router } from '@angular/router';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AUTH_TOKEN_KEY, authInterceptor } from './auth.interceptor';
import { AuthService } from '../services/auth.service';

function buildToken(expOffsetSeconds: number | null): string {
    const payload: Record<string, unknown> = { sub: 1, name: 'Ana', email: 'ana@ejemplo.com', role: 2 };
    if (expOffsetSeconds !== null) {
        payload['exp'] = Math.floor(Date.now() / 1000) + expOffsetSeconds;
    }
    return `header.${btoa(JSON.stringify(payload))}.signature`;
}

const VALID_TOKEN = () => buildToken(3600);
const EXPIRED_TOKEN = () => buildToken(-3600);

describe('authInterceptor', () => {
    let http: HttpClient;
    let httpMock: HttpTestingController;
    let auth: AuthService;
    let navigateSpy: ReturnType<typeof vi.spyOn>;

    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [
                provideRouter([]),
                provideHttpClient(withInterceptors([authInterceptor])),
                provideHttpClientTesting(),
            ],
        });
        http = TestBed.inject(HttpClient);
        httpMock = TestBed.inject(HttpTestingController);
        auth = TestBed.inject(AuthService);
        navigateSpy = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);
        localStorage.clear();
        auth.logout();
        navigateSpy.mockClear();
    });

    afterEach(() => {
        httpMock.verify();
        TestBed.resetTestingModule();
        localStorage.clear();
    });

    it('adjunta el Bearer en peticiones protegidas', () => {
        const token = VALID_TOKEN();
        localStorage.setItem(AUTH_TOKEN_KEY, token);

        http.get('http://localhost:8080/cdv-api/users/1').subscribe();

        const req = httpMock.expectOne('http://localhost:8080/cdv-api/users/1');
        expect(req.request.headers.get('Authorization')).toBe(`Bearer ${token}`);
        req.flush({});
    });

    it('no adjunta el Bearer en login', () => {
        localStorage.setItem(AUTH_TOKEN_KEY, VALID_TOKEN());

        http.post('http://localhost:8080/cdv-api/login', { email: 'a@a.com', password: 'x' }).subscribe();

        const req = httpMock.expectOne('http://localhost:8080/cdv-api/login');
        expect(req.request.headers.has('Authorization')).toBe(false);
        req.flush({});
    });

    it('no adjunta el Bearer en registro', () => {
        localStorage.setItem(AUTH_TOKEN_KEY, VALID_TOKEN());

        http.post('http://localhost:8080/cdv-api/users', { email: 'a@a.com' }).subscribe();

        const req = httpMock.expectOne('http://localhost:8080/cdv-api/users');
        expect(req.request.headers.has('Authorization')).toBe(false);
        req.flush({});
    });

    it('no adjunta nada si no hay token', () => {
        http.get('http://localhost:8080/cdv-api/cabins').subscribe();

        const req = httpMock.expectOne('http://localhost:8080/cdv-api/cabins');
        expect(req.request.headers.has('Authorization')).toBe(false);
        req.flush([]);
    });

    it('cierra la sesión sin llamar al backend si el token expiró', () => {
        localStorage.setItem(AUTH_TOKEN_KEY, EXPIRED_TOKEN());

        let status: number | undefined;
        http.get('http://localhost:8080/cdv-api/users/1').subscribe({
            error: (err: HttpErrorResponse) => (status = err.status),
        });

        httpMock.expectNone('http://localhost:8080/cdv-api/users/1');
        expect(status).toBe(401);
        expect(auth.isAuthenticated()).toBe(false);
        expect(localStorage.getItem(AUTH_TOKEN_KEY)).toBeNull();
        expect(navigateSpy).toHaveBeenCalledWith(['/login']);
    });

    it('cierra la sesión sin llamar al backend si el token es inválido', () => {
        localStorage.setItem(AUTH_TOKEN_KEY, 'token-con-formato-invalido');

        let status: number | undefined;
        http.get('http://localhost:8080/cdv-api/cabins').subscribe({
            error: (err: HttpErrorResponse) => (status = err.status),
        });

        httpMock.expectNone('http://localhost:8080/cdv-api/cabins');
        expect(status).toBe(401);
        expect(auth.isAuthenticated()).toBe(false);
        expect(localStorage.getItem(AUTH_TOKEN_KEY)).toBeNull();
    });

    it('cierra la sesión cuando el backend responde 401', () => {
        const token = VALID_TOKEN();
        localStorage.setItem(AUTH_TOKEN_KEY, token);

        let status: number | undefined;
        http.get('http://localhost:8080/cdv-api/users/1').subscribe({
            error: (err: HttpErrorResponse) => (status = err.status),
        });

        const req = httpMock.expectOne('http://localhost:8080/cdv-api/users/1');
        req.flush({ message: 'Token inválido o expirado' }, { status: 401, statusText: 'Unauthorized' });

        expect(status).toBe(401);
        expect(auth.isAuthenticated()).toBe(false);
        expect(localStorage.getItem(AUTH_TOKEN_KEY)).toBeNull();
        expect(navigateSpy).toHaveBeenCalledWith(['/login']);
    });

    it('no cierra la sesión cuando el backend responde 403', () => {
        const token = VALID_TOKEN();
        localStorage.setItem(AUTH_TOKEN_KEY, token);

        let status: number | undefined;
        http.get('http://localhost:8080/cdv-api/users/2').subscribe({
            error: (err: HttpErrorResponse) => (status = err.status),
        });

        const req = httpMock.expectOne('http://localhost:8080/cdv-api/users/2');
        req.flush({ message: 'No tiene permisos' }, { status: 403, statusText: 'Forbidden' });

        expect(status).toBe(403);
        expect(localStorage.getItem(AUTH_TOKEN_KEY)).toBe(token);
        expect(navigateSpy).not.toHaveBeenCalled();
    });
});
