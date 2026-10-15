import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
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
  guestId: 1,
  status: 'pendiente',
  createdAt: '2026-10-01T00:00:00Z',
  cancelledAt: null,
};

describe('ReservationCard', () => {
  let component: ReservationCard;
  let fixture: ComponentFixture<ReservationCard>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ReservationCard],
      providers: [
        provideRouter([]),
        {
          provide: ReservationService,
          useValue: {
            cancel: () => of({ message: 'ok', reservation }),
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

  it('permite consultar los pagos salvo cuando está cancelada', () => {
    expect(component.canViewPayments).toBe(true);

    component.reservation = { ...reservation, status: 'cancelada' };
    expect(component.canViewPayments).toBe(false);
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
