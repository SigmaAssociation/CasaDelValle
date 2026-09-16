import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';
import { isTokenExpired } from './token.utils';

export const AUTH_TOKEN_KEY = 'cdv_token';

function getStoredToken(): string | null {
    try {
        if (typeof localStorage === 'undefined') {
            return null;
        }
        return localStorage.getItem(AUTH_TOKEN_KEY);
    } catch {
        return null;
    }
}

export function isPublicAuthRequest(url: string, method: string): boolean {
    if (method !== 'POST') {
        return false;
    }
    if (url.endsWith('/cdv-api/login')) {
        return true;
    }
    if (/\/cdv-api\/users\/?$/.test(url)) {
        return true;
    }
    return false;
}

function closeSession(auth: AuthService, router: Router): void {
    auth.logout();
    if (!router.url.startsWith('/login')) {
        void router.navigate(['/login']).catch(() => undefined);
    }
}

function unauthorizedError(url: string): HttpErrorResponse {
    return new HttpErrorResponse({ status: 401, statusText: 'Unauthorized', url });
}

export const authInterceptor: HttpInterceptorFn = (req, next) => {
    if (isPublicAuthRequest(req.url, req.method)) {
        return next(req);
    }

    const auth = inject(AuthService);
    const router = inject(Router);

    const token = auth.getToken() ?? getStoredToken();
    if (!token) {
        return next(req);
    }

    if (isTokenExpired(token)) {
        closeSession(auth, router);
        return throwError(() => unauthorizedError(req.url));
    }

    const authReq = req.clone({
        setHeaders: {
            Authorization: `Bearer ${token}`,
        },
    });

    return next(authReq).pipe(
        catchError((err: unknown) => {
            if (err instanceof HttpErrorResponse && err.status === 401) {
                closeSession(auth, router);
            }
            return throwError(() => err);
        }),
    );
};
