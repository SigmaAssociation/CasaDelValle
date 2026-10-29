import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { IncomeReportResponse } from '../models/report-income';
import { ReportIncomeService } from './reports-income.service';

const emptyReport: IncomeReportResponse = {
  movements: [],
  total_payments: 0,
  total_refunds: 0,
  total_income: 0,
  total_movements: 0,
  start_date: '2026-10-01',
  end_date: '2026-10-31',
};

describe('ReportIncomeService', () => {
  let service: ReportIncomeService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ReportIncomeService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('consulta el reporte con start_date y end_date', () => {
    let result: IncomeReportResponse | undefined;
    service
      .getIncomeReport({ startDate: '2026-10-01', endDate: '2026-10-31' })
      .subscribe((response) => (result = response));

    const req = httpMock.expectOne((request) => request.url === '/cdv-api/reports/income');
    expect(req.request.method).toBe('GET');
    expect(req.request.params.get('start_date')).toBe('2026-10-01');
    expect(req.request.params.get('end_date')).toBe('2026-10-31');
    expect(req.request.params.has('cabin_id')).toBe(false);

    req.flush(emptyReport);
    expect(result).toEqual(emptyReport);
  });

  it('incluye cabin_id cuando se envía', () => {
    service
      .getIncomeReport({ startDate: '2026-10-01', endDate: '2026-10-31', cabinId: 4 })
      .subscribe();

    const req = httpMock.expectOne((request) => request.url === '/cdv-api/reports/income');
    expect(req.request.params.get('cabin_id')).toBe('4');
    req.flush(emptyReport);
  });
});
