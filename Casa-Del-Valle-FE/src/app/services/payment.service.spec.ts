import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { PaymentService } from './payment.service';
import { ReservationPayments } from '../models/reservation-payments';

describe('PaymentService', () => {
  let service: PaymentService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(PaymentService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('obtiene los pagos de una reservación por GET', () => {
    const body: ReservationPayments = {
      reservationId: 5,
      total: 1000,
      paid: 300,
      balance: 700,
      payments: [],
    };

    let result: ReservationPayments | undefined;
    service.getByReservation(5).subscribe((res) => (result = res));

    const req = httpMock.expectOne('/cdv-api/payments/reservation/5');
    expect(req.request.method).toBe('GET');
    req.flush(body);

    expect(result).toEqual(body);
  });

  it('registra un pago por POST con el cuerpo enviado', () => {
    const payload = { amount: 200, method: 'tarjeta' as const };

    service.create(5, payload).subscribe();

    const req = httpMock.expectOne('/cdv-api/payments/reservation/5');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(payload);
    req.flush({});
  });
});
