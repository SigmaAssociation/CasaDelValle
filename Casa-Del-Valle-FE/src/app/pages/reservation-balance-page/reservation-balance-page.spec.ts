import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ReservationBalancePage } from './reservation-balance-page';
import { ActivatedRoute, convertToParamMap } from '@angular/router';
import { of } from 'rxjs';

describe('ReservationBalancePage', () => {
  let component: ReservationBalancePage;
  let fixture: ComponentFixture<ReservationBalancePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ReservationBalancePage],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: {
            paramMap: of(convertToParamMap({ id: '1' })),
            params: of({ id: '1' }),
            snapshot: {
              paramMap: convertToParamMap({ id: '1' }),
              params: { id: '1' }
            }
          }
        }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(ReservationBalancePage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
