import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';
import { ReservationCard } from './reservation-card';
import { ReservationService } from '../../services/reservation.service';
import { PaymentService } from '../../services/payment.service';
import { AuthService } from '../../services/auth.service';

const reservation = {
  id: 1,
  cabinId: 1,
  cabinName: 'Cabaña Luna',
  startDate: '2026-10-10',
  endDate: '2026-10-15',
  totalPrice: 1500,
  paidTotal: 1500,
  guestId: 1,
  status: 'confirmada',
  createdAt: '2026-10-01T00:00:00Z',
  cancelledAt: null,
};

const checkinResponse = {
  message: 'Check-in registrado exitosamente',
  reservation: { ...reservation, status: 'en_curso' },
};

function textOf(fixture: ComponentFixture<ReservationCard>): string {
  return fixture.nativeElement.textContent as string;
}

function daysFromToday(days: number): string {
  const d = new Date();
  d.setDate(d.getDate() + days);
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${d.getFullYear()}-${month}-${day}`;
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
            payAnticipo: () => of({ message: 'ok', amount: 450, reservation }),
          },
        },
        {
          provide: PaymentService,
          useValue: {
            getByReservation: () =>
              of({ reservationId: 1, total: 1500, paid: 450, balance: 1050, payments: [] }),
            create: () => of({ message: 'ok' }),
          },
        },
        {
          provide: AuthService,
          useValue: {
            getCurrentUserId: () => 1,
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

  it('should show the check-in button for hosts within the stay dates', () => {
    component.reservation = {
      ...reservation,
      startDate: daysFromToday(-1),
      endDate: daysFromToday(2),
    };
    fixture.componentRef.setInput('showCheckin', true);
    fixture.detectChanges();
    expect(textOf(fixture)).toContain('Registrar check-in');
  });

  it('should hide the check-in button days before the stay', () => {
    component.reservation = {
      ...reservation,
      startDate: daysFromToday(5),
      endDate: daysFromToday(7),
    };
    fixture.componentRef.setInput('showCheckin', true);
    fixture.detectChanges();
    expect(textOf(fixture)).not.toContain('Registrar check-in');
    expect(textOf(fixture)).toContain('El check-in estará disponible');
  });

  it('should hide the check-in button days after the stay', () => {
    component.reservation = {
      ...reservation,
      startDate: daysFromToday(-7),
      endDate: daysFromToday(-5),
    };
    fixture.componentRef.setInput('showCheckin', true);
    fixture.detectChanges();
    expect(textOf(fixture)).not.toContain('Registrar check-in');
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

    fixture.componentRef.setInput('showCheckin', true);
    component.registerCheckin();
    fixture.detectChanges();

    expect(textOf(fixture)).toContain('El pago total de la reservación no está completo');
  });

  it('permite consultar los pagos salvo cuando está cancelada', () => {
    expect(component.canViewPayments).toBe(true);

    component.reservation = { ...reservation, status: 'cancelada' };
    expect(component.canViewPayments).toBe(false);
  });

  it('ofrece agregar pago solo con saldo pendiente en confirmada o en curso', () => {
    component.reservation = { ...reservation, status: 'confirmada', paidTotal: 450 };
    expect(component.canAddPayment).toBe(true);

    component.reservation = { ...reservation, status: 'confirmada', paidTotal: 1500 };
    expect(component.canAddPayment).toBe(false);

    component.reservation = { ...reservation, status: 'pendiente', paidTotal: 0 };
    expect(component.canAddPayment).toBe(false);

    component.reservation = { ...reservation, status: 'en_curso', paidTotal: 1000 };
    expect(component.canAddPayment).toBe(true);
  });

  it('muestra el botón Agregar pago cuando hay saldo', () => {
    fixture.componentRef.setInput('reservation', { ...reservation, status: 'confirmada', paidTotal: 450 });
    fixture.detectChanges();
    expect(textOf(fixture)).toContain('Agregar pago');
  });

  it('abre y cierra el modal de pagos', () => {
    component.openPaymentsModal();
    expect(component.showPaymentsModal()).toBe(true);

    component.closePaymentsModal();
    expect(component.showPaymentsModal()).toBe(false);
  });

  it('reenvía el evento paymentRegistered', () => {
    const response = { message: 'Pago registrado.' } as any;
    let emitted: any;
    component.paymentRegistered.subscribe((res) => (emitted = res));

    component.onPaymentRegistered(response);

    expect(emitted).toEqual(response);
  });
});
