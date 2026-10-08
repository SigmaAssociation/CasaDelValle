import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MyReservationsPage } from './my-reservations-page';

describe('MyReservationsPage', () => {
  let component: MyReservationsPage;
  let fixture: ComponentFixture<MyReservationsPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MyReservationsPage],
    }).compileComponents();

    fixture = TestBed.createComponent(MyReservationsPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
