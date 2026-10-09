import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
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
});
