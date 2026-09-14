/**
 * Devuelve true si el JWT está expirado o no tiene un formato válido.
 * Los tokens sin claim `exp` se consideran vigentes: en ese caso la
 * validez la decide el backend (un 401 dispara el cierre de sesión).
 */
export function isTokenExpired(token: string): boolean {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) {
            return true;
        }
        const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
        const json = JSON.parse(atob(base64));
        const exp = Number(json['exp']);
        if (!Number.isFinite(exp)) {
            return false;
        }
        return Math.floor(Date.now() / 1000) >= exp;
    } catch {
        return true;
    }
}
