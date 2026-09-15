import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { Router } from '@angular/router';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Login } from './login';
import { AuthService } from '../../services/auth.service';
import { LoginResponse } from '../../models/auth';

const validForm = {
    invalid: false,
    value: { email: 'usuario@ejemplo.com', password: 'Contraseña123!' },
} as unknown as Parameters<Login['onSubmit']>[0];

describe('Login', () => {
    let component: Login;
    let fixture: ComponentFixture<Login>;
    let loginSpy: ReturnType<typeof vi.fn>;
    let navigateSpy: ReturnType<typeof vi.fn>;

    beforeEach(async () => {
        loginSpy = vi.fn();
        navigateSpy = vi.fn();

        await TestBed.configureTestingModule({
            imports: [Login],
            providers: [
                { provide: AuthService, useValue: { login: loginSpy } },
                { provide: Router, useValue: { navigate: navigateSpy } },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(Login);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('no envía credenciales si el formulario es inválido', () => {
        component.onSubmit({ invalid: true, value: {} } as Parameters<Login['onSubmit']>[0]);
        expect(loginSpy).not.toHaveBeenCalled();
    });

    it('muestra mensaje de error cuando el servidor responde 401', () => {
        loginSpy.mockReturnValue(throwError(() => ({ status: 401 })));
        component.onSubmit(validForm);
        expect(component.errorMessage()).toContain('incorrectos');
    });

    it('muestra error de servidor cuando no hay conexión', () => {
        loginSpy.mockReturnValue(throwError(() => ({ status: 0 })));
        component.onSubmit(validForm);
        expect(component.errorMessage()).toContain('servidor');
    });

    it('muestra mensaje de éxito y redirige a la página principal', () => {
        vi.useFakeTimers();
        const response: LoginResponse = {
            message: 'Usuario autenticado exitosamente',
            token: 'jwt-123',
        };
        loginSpy.mockReturnValue(of(response));

        component.onSubmit(validForm);
        expect(component.successMessage()).toContain('¡Bienvenido');

        vi.advanceTimersByTime(1200);
        expect(navigateSpy).toHaveBeenCalledWith(['/']);
        vi.useRealTimers();
    });
});