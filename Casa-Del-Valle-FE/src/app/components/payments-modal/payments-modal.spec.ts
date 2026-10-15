import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { PaymentsModal } from './payments-modal';
import { PaymentService } from '../../services/payment.service';
import { Reservation } from '../../models/reservation';
import { ReservationPayments } from '../../models/reservation-payments';
import { CreatePaymentResponse } from '../../models/create-payment-response';

const reservation: Reservation = {
  id: 1,
  cabinId: 1,
  cabinName: 'Cabaña Luna',
  startDate: '2026-10-10',
  endDate: '2026-10-15',
  totalPrice: 1000,
  guestId: 1,
  status: 'confirmada',
  createdAt: '2026-10-01T00:00:00Z',
  cancelledAt: null,
};

const summary: ReservationPayments = {
  reservationId: 1,
  total: 1000,
  paid: 300,
  balance: 700,
  payments: [
    { id: 1, reservationId: 1, userId: 1, amount: 300, method: 'anticipo', date: '2026-10-02' },
  ],
};

describe('PaymentsModal', () => {
  let component: PaymentsModal;
  let fixture: ComponentFixture<PaymentsModal>;
  let getByReservation: ReturnType<typeof vi.fn>;
  let create: ReturnType<typeof vi.fn>;
  let paymentService: { getByReservation: typeof getByReservation; create: typeof create };

  beforeEach(() => {
    getByReservation = vi.fn().mockReturnValue(of({ ...summary }));
    create = vi.fn();
    paymentService = { getByReservation, create };
  });

  function createComponent(status = 'confirmada'): void {
    TestBed.configureTestingModule({
      imports: [PaymentsModal],
      providers: [{ provide: PaymentService, useValue: paymentService }],
    });
    fixture = TestBed.createComponent(PaymentsModal);
    component = fixture.componentInstance;
    component.reservation = { ...reservation, status };
    fixture.detectChanges();
  }

  it('should create', () => {
    createComponent();
    expect(component).toBeTruthy();
  });

  it('carga el resumen de pagos al iniciar', () => {
    createComponent();

    expect(getByReservation).toHaveBeenCalledWith(1);
    expect(component.isLoading()).toBe(false);
    expect(component.summary()?.balance).toBe(700);
    expect(component.summary()?.payments.length).toBe(1);
  });

  it('marca error de carga si falla el GET', () => {
    getByReservation.mockReturnValue(throwError(() => new Error('boom')));
    createComponent();

    expect(component.isLoading()).toBe(false);
    expect(component.loadError()).toBe(true);
  });

  it('solo permite registrar en reservaciones confirmadas o en curso', () => {
    createComponent('confirmada');
    expect(component.canRegister).toBe(true);

    component.reservation = { ...reservation, status: 'en_curso' };
    expect(component.canRegister).toBe(true);
  });

  it('no permite registrar en reservaciones pendientes, finalizadas o canceladas', () => {
    createComponent('pendiente');
    expect(component.canRegister).toBe(false);

    component.reservation = { ...reservation, status: 'finalizada' };
    expect(component.canRegister).toBe(false);

    component.reservation = { ...reservation, status: 'cancelada' };
    expect(component.canRegister).toBe(false);
  });

  it('rechaza montos menores o iguales a cero sin llamar al backend', () => {
    createComponent();

    component.amount.set(0);
    component.submit();

    expect(component.formError()).toBe('El monto debe ser mayor que 0.');
    expect(create).not.toHaveBeenCalled();
  });

  it('rechaza montos mayores al saldo pendiente sin llamar al backend', () => {
    createComponent();

    component.amount.set(800);
    component.submit();

    expect(component.formError()).toBe('El monto no puede superar el saldo pendiente.');
    expect(create).not.toHaveBeenCalled();
  });

  it('registra un pago válido, actualiza el resumen y emite el evento', () => {
    createComponent();
    const response: CreatePaymentResponse = {
      message: 'Pago registrado.',
      payment: {
        id: 2,
        reservationId: 1,
        userId: 1,
        amount: 200,
        method: 'tarjeta',
        date: '2026-10-05',
      },
      reservationId: 1,
      total: 1000,
      paid: 500,
      balance: 500,
      payments: summary.payments,
    };
    create.mockReturnValue(of(response));

    let emitted: CreatePaymentResponse | undefined;
    component.registered.subscribe((res) => (emitted = res));

    component.amount.set(200);
    component.method.set('tarjeta');
    component.submit();

    expect(create).toHaveBeenCalledWith(1, { amount: 200, method: 'tarjeta' });
    expect(component.summary()?.balance).toBe(500);
    expect(component.amount()).toBeNull();
    expect(emitted).toEqual(response);
  });

  it('muestra el mensaje del backend cuando el pago falla', () => {
    createComponent();
    create.mockReturnValue(
      throwError(() => ({ error: { message: 'El monto no puede superar el saldo pendiente.' } })),
    );

    component.amount.set(200);
    component.submit();

    expect(component.formError()).toBe('El monto no puede superar el saldo pendiente.');
  });

  it('emite dismiss al cerrar', () => {
    createComponent();
    let dismissed = false;
    component.dismiss.subscribe(() => (dismissed = true));

    component.close();

    expect(dismissed).toBe(true);
  });
});
