import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { catchError, map, of } from 'rxjs';
import { AuthService } from '../services/auth.service';
import { CabinService } from '../services/cabin.service';


export const cabinOwnerGuard: CanActivateFn = (route) => {
  const auth = inject(AuthService);
  const cabinService = inject(CabinService);
  const router = inject(Router);

  const id = Number(route.paramMap.get('id'));
  if (!Number.isInteger(id) || id <= 0) {
    return router.createUrlTree(['/']);
  }

  if (auth.user()?.role === 1) {
    return true;
  }

  const currentUserId = auth.getCurrentUserId();
  if (!currentUserId) {
    return router.createUrlTree(['/login']);
  }

  return cabinService.getCabinById(id).pipe(
    map((cabin) =>
      cabin?.host_id === currentUserId ? true : router.createUrlTree(['/']),
    ),
    catchError(() => of(router.createUrlTree(['/']))),
  );
};
