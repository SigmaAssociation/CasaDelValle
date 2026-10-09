import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { ReservationForm } from './reservation-form';
import { AuthService } from '../../services/auth.service';

function toDateString(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function daysFromToday(days: number): string {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return toDateString(date);
}

describe('ReservationForm', () => {
  let component: ReservationForm;
  let fixture: ComponentFixture<ReservationForm>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ReservationForm],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: AuthService, useValue: { getCurrentUserId: () => 7 } },
      ],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(ReservationForm);
    component = fixture.componentInstance;
    component.cabinId = 3;
    component.nightlyPrice = 150;
    fixture.detectChanges();
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('marca el formulario como inválido sin fechas', () => {
    expect(component.reservationForm.valid).toBe(false);
    component.submit();
    expect(component.reservationForm.get('start_date')?.touched).toBe(true);
    expect(component.reservationForm.get('end_date')?.touched).toBe(true);
  });

  it('rechaza una fecha de inicio anterior a hoy', () => {
    component.reservationForm.patchValue({
      start_date: daysFromToday(-1),
      end_date: daysFromToday(3),
    });
    expect(component.reservationForm.errors?.['startInPast']).toBe(true);
    expect(component.startRangeError).toBe('La fecha de inicio no puede ser anterior a hoy.');
  });

  it('rechaza una fecha de fin igual o anterior a la de inicio', () => {
    component.reservationForm.patchValue({
      start_date: daysFromToday(2),
      end_date: daysFromToday(2),
    });
    expect(component.reservationForm.errors?.['endBeforeStart']).toBe(true);

    component.reservationForm.patchValue({ end_date: daysFromToday(1) });
    expect(component.reservationForm.errors?.['endBeforeStart']).toBe(true);
  });

  it('acepta un rango de fechas válido y calcula noches y total', () => {
    component.reservationForm.patchValue({
      start_date: daysFromToday(1),
      end_date: daysFromToday(4),
    });
    expect(component.reservationForm.valid).toBe(true);
    expect(component.nights).toBe(3);
    expect(component.estimatedTotal).toBe(450);
  });

  it('envía user_id y cabin_id automáticamente junto con las fechas', () => {
    const startDate = daysFromToday(1);
    const endDate = daysFromToday(2);

    component.reservationForm.patchValue({
      start_date: startDate,
      end_date: endDate,
    });
    component.submit();

    const req = httpMock.expectOne('/cdv-api/reservations');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({
      user_id: 7,
      cabin_id: 3,
      start_date: startDate,
      end_date: endDate,
    });

    req.flush({ message: 'Reservación registrada exitosamente', reservation_id: 10 });
    expect(component.isSuccess()).toBe(true);
    expect(component.reservationId()).toBe(10);
    expect(component.isSubmitting()).toBe(false);
  });

  it('muestra el mensaje del backend cuando hay conflicto de fechas', () => {
    component.reservationForm.patchValue({
      start_date: daysFromToday(1),
      end_date: daysFromToday(2),
    });
    component.submit();

    const req = httpMock.expectOne('/cdv-api/reservations');
    req.flush(
      { message: 'La cabaña no está disponible en el rango de fechas seleccionado' },
      { status: 409, statusText: 'Conflict' }
    );

    expect(component.isError()).toBe(true);
    expect(component.errorMessage()).toBe('La cabaña no está disponible en el rango de fechas seleccionado');
    expect(component.isSuccess()).toBe(false);
  });
});
