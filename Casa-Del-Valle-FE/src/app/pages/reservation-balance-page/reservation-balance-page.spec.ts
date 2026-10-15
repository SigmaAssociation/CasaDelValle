import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReservationBalancePage } from './reservation-balance-page';

describe('ReservationBalancePage', () => {
  let component: ReservationBalancePage;
  let fixture: ComponentFixture<ReservationBalancePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ReservationBalancePage],
    }).compileComponents();

    fixture = TestBed.createComponent(ReservationBalancePage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
