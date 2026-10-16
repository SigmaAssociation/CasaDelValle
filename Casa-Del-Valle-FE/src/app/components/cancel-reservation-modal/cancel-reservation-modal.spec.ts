import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CancelReservationModal } from './cancel-reservation-modal';
import { Reservation } from '../../models/reservation';

function inDays(days: number): string {
  const d = new Date();
  d.setHours(0, 0, 0, 0);
  d.setDate(d.getDate() + days);
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${d.getFullYear()}-${m}-${day}`;
}

function makeReservation(overrides: Partial<Reservation> = {}): Reservation {
  return {
    id: 1,
    cabinId: 1,
    cabinName: 'Cabaña Luna',
    startDate: inDays(10),
    endDate: inDays(15),
    totalPrice: 1500,
    status: 'confirmada',
    createdAt: '2026-10-01T00:00:00Z',
    ...overrides,
  };
}

describe('CancelReservationModal', () => {
  let component: CancelReservationModal;
  let fixture: ComponentFixture<CancelReservationModal>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CancelReservationModal],
    }).compileComponents();

    fixture = TestBed.createComponent(CancelReservationModal);
    component = fixture.componentInstance;
    component.reservation = makeReservation();
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('muestra el monto total del reembolso con 10 días de anticipación', () => {
    component.reservation = makeReservation({ startDate: inDays(10) });
    fixture.detectChanges();

    expect(component.refundTier).toBe('total');
    expect(component.refundAmount).toBe(450);
    expect(component.refundMessage).toContain('Q 450.00');
  });

  it('muestra el monto parcial (70%) con 5 días de anticipación', () => {
    component.reservation = makeReservation({ startDate: inDays(5) });
    fixture.detectChanges();

    expect(component.refundTier).toBe('parcial');
    expect(component.refundAmount).toBe(315);
    expect(component.refundMessage).toContain('Q 315.00');
  });

  it('avisa que no aplica reembolso con menos de 3 días', () => {
    component.reservation = makeReservation({ startDate: inDays(1) });
    fixture.detectChanges();

    expect(component.refundTier).toBe('ninguno');
    expect(component.hasRefund).toBe(false);
    expect(component.refundMessage).toContain('no aplica reembolso');
  });

  it('avisa que no hay pagos en una reservación pendiente', () => {
    component.reservation = makeReservation({ startDate: inDays(10), status: 'pendiente' });
    fixture.detectChanges();

    expect(component.refundAmount).toBe(0);
    expect(component.refundMessage).toContain('no tiene pagos registrados');
  });
});
