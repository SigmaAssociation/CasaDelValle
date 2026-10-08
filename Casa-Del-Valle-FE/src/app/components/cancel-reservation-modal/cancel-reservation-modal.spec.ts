import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CancelReservationModal } from './cancel-reservation-modal';

describe('CancelReservationModal', () => {
  let component: CancelReservationModal;
  let fixture: ComponentFixture<CancelReservationModal>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CancelReservationModal],
    }).compileComponents();

    fixture = TestBed.createComponent(CancelReservationModal);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
