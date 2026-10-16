import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FinancialHistoryModal } from './financial-history-modal';
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
  movements: [
    { id: 1, type: 'pago', subtype: 'anticipo', amount: 300, method: 'CASH', createdAt: '2026-10-01T10:00:00Z' },
    { id: 2, type: 'reembolso', subtype: 'reembolso', amount: 210, method: 'TRANSFER_SIMULATED', createdAt: '2026-10-02T10:00:00Z' },
  ],
  totalPaid: 300,
  totalRefunded: 210,
  refundableAvailable: 90,
};

describe('FinancialHistoryModal', () => {
  let component: FinancialHistoryModal;
  let fixture: ComponentFixture<FinancialHistoryModal>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FinancialHistoryModal],
    }).compileComponents();

    fixture = TestBed.createComponent(FinancialHistoryModal);
    component = fixture.componentInstance;
    component.reservation = reservation;
    component.history = history;
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('labels payments and refunds', () => {
    expect(component.movementLabel(history.movements[0])).toBe('Anticipo');
    expect(component.movementLabel(history.movements[1])).toBe('Reembolso');
  });

  it('renders the totals', () => {
    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('300.00');
    expect(text).toContain('210.00');
    expect(text).toContain('90.00');
  });
});
