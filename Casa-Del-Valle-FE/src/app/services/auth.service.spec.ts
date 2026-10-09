import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import {
    HttpTestingController,
    provideHttpClientTesting,
} from '@angular/common/http/testing';
import { beforeEach, describe, expect, it } from 'vitest';
import { AuthService } from './auth.service';

const API_URL = '/cdv-api/login';

function buildToken(payload: Record<string, unknown>): string {
    const encoded = btoa(JSON.stringify(payload));
    return `header.${encoded}.signature`;
}

describe('AuthService', () => {
    function createService() {
        TestBed.resetTestingModule();
        TestBed.configureTestingModule({
            providers: [provideHttpClient(), provideHttpClientTesting()],
        });
        return {
            service: TestBed.inject(AuthService),
            httpMock: TestBed.inject(HttpTestingController),
        };
    }

    it('almacena el JWT cuando el login es exitoso', () => {
        localStorage.clear();
        const { service, httpMock } = createService();
        const token = buildToken({
            sub: 1,
            name: 'Ana',
            email: 'ana@ejemplo.com',
            role: 2,
        });

        service.login('ana@ejemplo.com', 'Contraseña123!').subscribe((res) => {
            expect(res.token).toBe(token);
        });

        const req = httpMock.expectOne(API_URL);
        expect(req.request.method).toBe('POST');
        expect(req.request.body).toEqual({ email: 'ana@ejemplo.com', password: 'Contraseña123!' });

        req.flush({ message: 'Usuario autenticado exitosamente', token });

        expect(service.getToken()).toBe(token);
        expect(service.isAuthenticated()).toBe(true);
        expect(service.user()?.email).toBe('ana@ejemplo.com');
        expect(localStorage.getItem('cdv_token')).toBe(token);
        httpMock.verify();
    });

    it('no autentica cuando el backend responde 401', () => {
        localStorage.clear();
        const { service, httpMock } = createService();
        service.login('ana@ejemplo.com', 'mala').subscribe({
            error: (err: { status: number }) => expect(err.status).toBe(401),
        });

        const req = httpMock.expectOne(API_URL);
        req.flush(
            { message: 'Correo o contraseña incorrectos', token: '' },
            { status: 401, statusText: 'Unauthorized' },
        );

        expect(service.isAuthenticated()).toBe(false);
        expect(service.getToken()).toBeNull();
        httpMock.verify();
    });

    it('no autentica si el backend responde 200 sin token', () => {
        localStorage.clear();
        const { service, httpMock } = createService();
        service.login('ana@ejemplo.com', 'Contraseña123!').subscribe();

        httpMock.expectOne(API_URL).flush({ message: 'Error inesperado', token: '' });

        expect(service.isAuthenticated()).toBe(false);
        expect(service.getToken()).toBeNull();
        httpMock.verify();
    });

    it('logout limpia el token del almacenamiento', () => {
        localStorage.clear();
        const { service, httpMock } = createService();
        const token = buildToken({ sub: 1, name: 'Ana', email: 'ana@ejemplo.com', role: 2 });

        service.login('ana@ejemplo.com', 'Contraseña123!').subscribe();
        httpMock.expectOne(API_URL).flush({ message: 'ok', token });

        service.logout();
        expect(service.isAuthenticated()).toBe(false);
        expect(service.getToken()).toBeNull();
        expect(localStorage.getItem('cdv_token')).toBeNull();
        httpMock.verify();
    });

    it('cierra la sesión al iniciar si el token guardado expiró', () => {
        localStorage.clear();
        const expired = buildToken({ sub: 1, exp: Math.floor(Date.now() / 1000) - 10 });
        localStorage.setItem('cdv_token', expired);
        const { service, httpMock } = createService();

        expect(service.isAuthenticated()).toBe(false);
        expect(service.getToken()).toBeNull();
        expect(localStorage.getItem('cdv_token')).toBeNull();
        httpMock.verify();
    });
});