import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';
import { ReservationCard } from './reservation-card';
import { ReservationService } from '../../services/reservation.service';

const reservation = {
  id: 1,
  cabinId: 1,
  cabinName: 'Cabaña Luna',
  startDate: '2026-10-10',
  endDate: '2026-10-15',
  totalPrice: 1500,
  status: 'activa',
  createdAt: '2026-10-01T00:00:00Z',
  cancelledAt: null,
};

const checkinResponse = {
  message: 'Check-in registrado exitosamente',
  reservation: { ...reservation, status: 'checkin' },
};

function textOf(fixture: ComponentFixture<ReservationCard>): string {
  return fixture.nativeElement.textContent as string;
}

describe('ReservationCard', () => {
  let component: ReservationCard;
  let fixture: ComponentFixture<ReservationCard>;
  let confirmSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(async () => {
    confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    await TestBed.configureTestingModule({
      imports: [ReservationCard],
      providers: [
        provideRouter([]),
        {
          provide: ReservationService,
          useValue: {
            cancel: () => of({ message: 'ok', reservation }),
            checkin: () => of(checkinResponse),
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ReservationCard);
    component = fixture.componentInstance;
    component.reservation = reservation;
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should hide the check-in button by default', () => {
    expect(textOf(fixture)).not.toContain('Registrar check-in');
  });

  it('should show the check-in button for hosts on active reservations', () => {
    fixture.componentRef.setInput('showCheckin', true);
    fixture.detectChanges();
    expect(textOf(fixture)).toContain('Registrar check-in');
  });

  it('should emit checkedIn after confirming', () => {
    fixture.componentRef.setInput('showCheckin', true);
    let emitted: unknown = null;
    component.checkedIn.subscribe((res) => (emitted = res));

    component.registerCheckin();

    expect(confirmSpy).toHaveBeenCalled();
    expect(emitted).toEqual(checkinResponse);
    expect(component.isCheckingIn()).toBe(false);
  });

  it('should surface the backend error when check-in is rejected', () => {
    const failingService = TestBed.inject(ReservationService) as unknown as {
      checkin: () => unknown;
    };
    failingService.checkin = () =>
      throwError(() => ({ error: { message: 'El pago total de la reservación no está completo' } }));

    component.showCheckin = true;
    component.registerCheckin();
    fixture.detectChanges();

    expect(textOf(fixture)).toContain('El pago total de la reservación no está completo');
  });
});
