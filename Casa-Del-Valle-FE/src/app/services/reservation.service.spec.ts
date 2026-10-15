import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { firstValueFrom } from 'rxjs';
import { ReservationService } from './reservation.service';

describe('ReservationService check-in', () => {
  let service: ReservationService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ReservationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should register the check-in with PATCH and cash amount', () => {
    const detail = {
      id: 10,
      user_id: 1,
      cabin_id: 5,
      host_id: 2,
      start_date: '2026-11-10',
      end_date: '2026-11-12',
      status: 'en_curso',
    };

    service.checkin(10, 200).subscribe((res) => {
      expect(res.message).toBe('Check-in registrado exitosamente');
      expect(res.reservation.status).toBe('en_curso');
    });

    const req = httpMock.expectOne('/cdv-api/reservations/10/checkin');
    expect(req.request.method).toBe('PATCH');
    expect(req.request.body).toEqual({ cash_amount: 200 });
    req.flush({ message: 'Check-in registrado exitosamente', reservation: detail });
  });

  it('should propagate the rejection when requirements are not met', async () => {
    const pending = firstValueFrom(service.checkin(11));

    const req = httpMock.expectOne('/cdv-api/reservations/11/checkin');
    req.flush(
      { message: 'El pago total de la reservación no está completo' },
      { status: 409, statusText: 'Conflict' }
    );

    await expect(pending).rejects.toMatchObject({ status: 409 });
  });
});
