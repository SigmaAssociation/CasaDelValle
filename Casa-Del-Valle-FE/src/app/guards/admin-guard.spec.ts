import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, Router, RouterStateSnapshot, UrlTree, provideRouter } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { adminGuard } from './admin-guard';

describe('adminGuard', () => {
  const route = {} as ActivatedRouteSnapshot;
  const state = {} as RouterStateSnapshot;

  function setup(isAdmin: boolean): void {
    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: { isAdmin: () => isAdmin } },
      ],
    });
  }

  it('allows administrators', () => {
    setup(true);
    const result = TestBed.runInInjectionContext(() => adminGuard(route, state));
    expect(result).toBe(true);
  });

  it('redirects non-administrators to the home page', () => {
    setup(false);
    const result = TestBed.runInInjectionContext(() => adminGuard(route, state));
    expect(result instanceof UrlTree).toBe(true);
    expect(TestBed.inject(Router).serializeUrl(result as UrlTree)).toBe('/');
  });
});
