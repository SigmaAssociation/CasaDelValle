import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { ActivatedRoute, convertToParamMap, provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { vi } from 'vitest';
import { CabinDetailPage } from './cabin-detail-page';
import { CabinService } from '../../services/cabin.service';
import { ReservationService } from '../../services/reservation.service';
import { AuthService } from '../../services/auth.service';
import { MaintenanceService } from '../../services/maintenance.service';

describe('CabinDetailPage maintenance access', () => {
  const user = signal({ id: 7, role: 3 });
  let fixture: ComponentFixture<CabinDetailPage>;
  beforeEach(async () => {
    user.set({ id: 7, role: 3 });
    await TestBed.configureTestingModule({
      imports: [CabinDetailPage],
      providers: [
        provideRouter([]),
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: convertToParamMap({id: '100'}) } } },
        { provide: AuthService, useValue: { user, getCurrentUserId: () => user().id, isHostOrAdmin: () => [1, 3].includes(user().role) } },
        { provide: CabinService, useValue: {
          getCabinById: () => of({id: 100, host_id: 7, name: 'Cabaña de prueba', price: 400, capacity: 4, address: 'Guatemala'}),
          getImagesByCabin: () => of([]),
        } },
        { provide: ReservationService, useValue: { getReservationsByCabin: () => of([]) } },
        { provide: MaintenanceService, useValue: { list: () => of([]) } },
      ],
    }).compileComponents();
  });

  function render() {
    fixture = TestBed.createComponent(CabinDetailPage);
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  }

  it('shows the owner a visible action that leads to the maintenance form', () => {
    const page = render();
    const action = page.querySelector<HTMLButtonElement>('button[aria-controls="cabin-maintenance-section"]');
    const section = page.querySelector<HTMLElement>('#cabin-maintenance-section');
    expect(action?.textContent).toContain('Poner en mantenimiento');
    expect(section?.querySelector('app-cabin-maintenance')).not.toBeNull();
    const scroll = vi.fn(); Object.defineProperty(section, 'scrollIntoView', { value: scroll });
    action!.click();
    expect(scroll).toHaveBeenCalledWith({behavior: 'smooth', block: 'start'});
  });

  it('shows the maintenance action and form when the owner has the administrator role', () => {
    user.set({id: 7, role: 1});
    const page = render();
    expect(page.querySelector('button[aria-controls="cabin-maintenance-section"]')).not.toBeNull();
    expect(page.querySelector('app-cabin-maintenance')).not.toBeNull();
  });

  it('does not offer maintenance to an administrator who does not own the cabin', () => {
    user.set({id: 1, role: 1});
    const page = render();
    expect(page.querySelector('button[aria-controls="cabin-maintenance-section"]')).toBeNull();
    expect(page.querySelector('app-cabin-maintenance')).toBeNull();
  });

  it('does not offer maintenance to another host', () => {
    user.set({id: 8, role: 3});
    const page = render();
    expect(page.querySelector('button[aria-controls="cabin-maintenance-section"]')).toBeNull();
    expect(page.querySelector('app-cabin-maintenance')).toBeNull();
  });

  it('does not offer maintenance to a guest', () => {
    user.set({id: 9, role: 2});
    const page = render();
    expect(page.querySelector('button[aria-controls="cabin-maintenance-section"]')).toBeNull();
    expect(page.querySelector('app-cabin-maintenance')).toBeNull();
  });
});
