import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RefundModal } from './refund-modal';
import { Reservation } from '../../models/reservation';
import { RefundsHistoryResponse } from '../../models/refund';

const reservation: Reservation = {
  id: 1,
  cabinId: 1,
  hostId: 2,
  cabinName: 'Cabaña Luna',
  startDate: '2026-10-10',
  endDate: '2026-10-15',
  totalPrice: 1500,
  guestId: 1,
  status: 'cancelada',
  createdAt: '2026-10-01T00:00:00Z',
};

const history: RefundsHistoryResponse = {
  movements: [],
  totalPaid: 300,
  totalRefunded: 60,
  refundableAvailable: 240,
};

describe('RefundModal', () => {
  let component: RefundModal;
  let fixture: ComponentFixture<RefundModal>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RefundModal],
    }).compileComponents();

    fixture = TestBed.createComponent(RefundModal);
    component = fixture.componentInstance;
    component.reservation = reservation;
    component.history = history;
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('defaults to a full refund of the available amount', () => {
    expect(component.mode()).toBe('total');
    expect(component.amountToRefund()).toBe(240);
    expect(component.canConfirm).toBe(true);
  });

  it('rejects a partial amount greater than the available', () => {
    component.mode.set('parcial');
    component.partialAmount.set(300);

    expect(component.validationError).not.toBeNull();
    expect(component.canConfirm).toBe(false);
  });

  it('accepts a partial amount within the available', () => {
    component.mode.set('parcial');
    component.partialAmount.set(100);

    expect(component.validationError).toBeNull();
    expect(component.amountToRefund()).toBe(100);
    expect(component.canConfirm).toBe(true);
  });

  it('emits the amount when confirming a partial refund', () => {
    component.mode.set('parcial');
    component.partialAmount.set(100);
    let emitted: number | undefined | null = null;
    component.confirm.subscribe((value) => (emitted = value));

    component.submit();

    expect(emitted).toBe(100);
  });

  it('emits undefined for a full refund', () => {
    let emitted: number | undefined | null = null;
    component.confirm.subscribe((value) => (emitted = value));

    component.submit();

    expect(emitted).toBeUndefined();
  });
});
