import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ReservationService } from './reservation.service';
import { RestConstants } from '../components/rest-constants';
import { CheckoutReservationResponse } from '../models/checkout-reservation-response';

describe('ReservationService.checkout', () => {
  let service: ReservationService;
  let http: HttpTestingController;
  const api = new RestConstants().getApiURL();

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ReservationService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('envía PATCH a reservations/{id}/checkout con cuerpo vacío', () => {
    service.checkout(7).subscribe();

    const req = http.expectOne(`${api}reservations/7/checkout`);
    expect(req.request.method).toBe('PATCH');
    expect(req.request.body).toEqual({});
    req.flush({});
  });

  it('devuelve la respuesta del backend', () => {
    const body = {
      message: 'Check-out registrado exitosamente',
      reservation: { id: 7, status: 'finalizada' },
    } as unknown as CheckoutReservationResponse;
    let result: CheckoutReservationResponse | undefined;

    service.checkout(7).subscribe((res) => (result = res));
    http.expectOne(`${api}reservations/7/checkout`).flush(body);

    expect(result?.message).toBe('Check-out registrado exitosamente');
    expect(result?.reservation.status).toBe('finalizada');
  });

  it('propaga el error 409 cuando no hay check-in previo', () => {
    let status = 0;

    service.checkout(7).subscribe({ error: (err) => (status = err.status) });
    http
      .expectOne(`${api}reservations/7/checkout`)
      .flush(
        { error: 'No se puede registrar el check-out sin un check-in previo' },
        { status: 409, statusText: 'Conflict' },
      );

    expect(status).toBe(409);
  });
});
