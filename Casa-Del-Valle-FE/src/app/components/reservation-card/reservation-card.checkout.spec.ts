import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Observable, of, throwError } from 'rxjs';
import { ReservationCard } from './reservation-card';
import { ReservationService } from '../../services/reservation.service';
import { AuthService } from '../../services/auth.service';
import { Reservation } from '../../models/reservation';
import { CheckoutReservationResponse } from '../../models/checkout-reservation-response';

const baseReservation: Reservation = {
  id: 15,
  cabinId: 3,
  cabinName: 'Cabaña del Lago',
  guestId: 99,
  guestName: 'Ana López',
  startDate: '2026-10-14',
  endDate: '2026-10-18',
  totalPrice: 1000,
  paidTotal: 1000,
  status: 'en_curso',
  createdAt: '2026-09-01T10:00:00',
};

describe('ReservationCard · check-out', () => {
  let fixture: ComponentFixture<ReservationCard>;
  let component: ReservationCard;
  let checkoutCalls: number[];
  let checkoutResult: () => Observable<CheckoutReservationResponse>;
  let originalConfirm: typeof window.confirm;
  let confirmAnswer: boolean;

  const okResponse = {
    message: 'Check-out registrado exitosamente',
    reservation: { id: 15, status: 'finalizada' },
  } as unknown as CheckoutReservationResponse;

  function setup(overrides: Partial<Reservation> = {}, showCheckout = true): void {
    fixture = TestBed.createComponent(ReservationCard);
    component = fixture.componentInstance;
    component.reservation = { ...baseReservation, ...overrides };
    fixture.componentRef.setInput('showCheckout', showCheckout);
    fixture.detectChanges();
  }

  function checkoutButton(): HTMLButtonElement | undefined {
    const buttons = Array.from(
      (fixture.nativeElement as HTMLElement).querySelectorAll('button'),
    ) as HTMLButtonElement[];
    return buttons.find((b) => (b.textContent ?? '').includes('check-out') || (b.textContent ?? '').includes('Registrando'));
  }

  beforeEach(() => {
    checkoutCalls = [];
    checkoutResult = () => of(okResponse);
    confirmAnswer = true;
    originalConfirm = window.confirm;
    window.confirm = () => confirmAnswer;

    TestBed.configureTestingModule({
      imports: [ReservationCard],
      providers: [
        provideRouter([]),
        {
          provide: ReservationService,
          useValue: {
            checkout: (id: number) => {
              checkoutCalls.push(id);
              return checkoutResult();
            },
          },
        },
        { provide: AuthService, useValue: { getCurrentUserId: () => 1 } },
      ],
    });
  });

  afterEach(() => {
    window.confirm = originalConfirm;
  });

  it('muestra el botón de check-out para una reservación en curso cuando showCheckout es true', () => {
    setup();
    expect(component.canCheckout).toBe(true);
    expect(checkoutButton()).toBeDefined();
  });

  it('no muestra el botón si showCheckout es false', () => {
    setup({}, false);
    expect(component.canCheckout).toBe(false);
    expect(checkoutButton()).toBeUndefined();
  });

  for (const status of ['pendiente', 'confirmada', 'finalizada', 'cancelada']) {
    it(`no muestra el botón cuando la reservación está ${status}`, () => {
      setup({ status });
      expect(component.canCheckout).toBe(false);
      expect(checkoutButton()).toBeUndefined();
    });
  }

  it('registra el check-out y emite checkedOut con la respuesta', () => {
    setup();
    let emitted: CheckoutReservationResponse | undefined;
    component.checkedOut.subscribe((res) => (emitted = res));

    checkoutButton()!.click();

    expect(checkoutCalls).toEqual([15]);
    expect(emitted).toBe(okResponse);
    expect(component.isCheckingOut()).toBe(false);
    expect(component.errorMessage()).toBeNull();
  });

  it('no llama al backend si el usuario cancela la confirmación', () => {
    setup();
    confirmAnswer = false;
    let emitted = false;
    component.checkedOut.subscribe(() => (emitted = true));

    component.registerCheckout();

    expect(checkoutCalls).toEqual([]);
    expect(emitted).toBe(false);
  });

  it('no llama al backend si la reservación no está en curso', () => {
    setup({ status: 'confirmada' });

    component.registerCheckout();

    expect(checkoutCalls).toEqual([]);
  });

  it('muestra el mensaje del backend cuando falta el check-in (409)', () => {
    setup();
    checkoutResult = () =>
      throwError(() => ({
        status: 409,
        error: { error: 'No se puede registrar el check-out sin un check-in previo' },
      }));
    let emitted = false;
    component.checkedOut.subscribe(() => (emitted = true));

    component.registerCheckout();
    fixture.detectChanges();

    expect(emitted).toBe(false);
    expect(component.isCheckingOut()).toBe(false);
    expect(component.errorMessage()).toBe('No se puede registrar el check-out sin un check-in previo');
    expect((fixture.nativeElement as HTMLElement).querySelector('[role="alert"]')?.textContent).toContain(
      'sin un check-in previo',
    );
  });

  it('usa un mensaje genérico si el error no trae detalle', () => {
    setup();
    checkoutResult = () => throwError(() => ({ status: 500, error: null }));

    component.registerCheckout();

    expect(component.errorMessage()).toBe('No se pudo registrar el check-out. Intenta de nuevo.');
  });

  it('deshabilita el botón y muestra "Registrando…" mientras se procesa', () => {
    setup();
    component.isCheckingOut.set(true);
    fixture.detectChanges();

    expect(component.canCheckout).toBe(false);
    const btn = checkoutButton()!;
    expect(btn.disabled).toBe(true);
    expect(btn.textContent).toContain('Registrando');
  });
});