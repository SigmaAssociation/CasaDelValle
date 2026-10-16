import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { RefundService } from './refund.service';
import { RefundsHistoryResponse } from '../models/refund';

const history: RefundsHistoryResponse = {
  movements: [],
  totalPaid: 300,
  totalRefunded: 0,
  refundableAvailable: 300,
};

describe('RefundService', () => {
  let service: RefundService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(RefundService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('gets the financial history of a reservation', () => {
    service.getHistory(5).subscribe((res) => {
      expect(res.refundableAvailable).toBe(300);
    });

    const req = httpMock.expectOne('/cdv-api/refunds/reservation/5');
    expect(req.request.method).toBe('GET');
    req.flush(history);
  });

  it('registers a partial refund with the given amount', () => {
    service.create(5, { amount: 120 }).subscribe((res) => {
      expect(res.message).toBe('ok');
    });

    const req = httpMock.expectOne('/cdv-api/refunds/reservation/5');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ amount: 120 });
    req.flush({ message: 'ok', refund: { id: 1, reservationId: 5, hostId: 2, amount: 120, method: 'TRANSFER_SIMULATED', createdAt: '' } });
  });

  it('registers a full refund when no amount is given', () => {
    service.create(5).subscribe();

    const req = httpMock.expectOne('/cdv-api/refunds/reservation/5');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({});
    req.flush({ message: 'ok', refund: {} });
  });
});
