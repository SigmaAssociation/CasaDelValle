import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CancelReservationModal } from './cancel-reservation-modal';

const reservation = {
  id: 1,
  cabinId: 1,
  cabinName: 'Cabaña Luna',
  startDate: '2026-10-10',
  endDate: '2026-10-15',
  totalPrice: 1500,
  status: 'activa',
  createdAt: '2026-10-01T00:00:00Z',
};

describe('CancelReservationModal', () => {
  let component: CancelReservationModal;
  let fixture: ComponentFixture<CancelReservationModal>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CancelReservationModal],
    }).compileComponents();

    fixture = TestBed.createComponent(CancelReservationModal);
    component = fixture.componentInstance;
    component.reservation = reservation;
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
