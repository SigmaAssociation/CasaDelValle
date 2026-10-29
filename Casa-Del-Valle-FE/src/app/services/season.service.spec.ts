import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { SeasonService } from './season.service';

describe('SeasonService', () => {
  let service: SeasonService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting()] });
    service = TestBed.inject(SeasonService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('creates a custom cabin seasonal price', () => {
    service.createCabinPrice(12, { startDate: '2026-12-20', endDate: '2026-12-31', adjustmentPercent: 25, fixedPrice: 650 }).subscribe();
    const request = http.expectOne('/cdv-api/cabin-season-prices/cabin/12');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ startDate: '2026-12-20', endDate: '2026-12-31', adjustmentPercent: 25, fixedPrice: 650 });
    request.flush({ id: 1 });
  });

  it('lists global seasons and cabin prices', () => {
    service.list().subscribe();
    service.listCabinPrices(12).subscribe();
    http.expectOne('/cdv-api/seasons').flush([]);
    http.expectOne('/cdv-api/cabin-season-prices/cabin/12').flush([]);
  });

  it('requests inactive seasons when includeInactive is true', () => {
    service.list(true).subscribe();
    http.expectOne('/cdv-api/seasons?all=true').flush([]);
  });

  it('creates a global season', () => {
    const body = {
      name: 'Navidad',
      suggestedAdjustmentPercent: 25,
      active: true,
      startDate: '2026-12-20',
      endDate: '2026-12-31',
    };
    service.create(body).subscribe();
    const request = http.expectOne('/cdv-api/seasons');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(body);
    request.flush({ id: 1, ...body });
  });

  it('updates a global season', () => {
    const body = {
      name: 'Navidad',
      suggestedAdjustmentPercent: 30,
      active: true,
      startDate: '2026-12-20',
      endDate: '2026-12-31',
    };
    service.update(5, body).subscribe();
    const request = http.expectOne('/cdv-api/seasons/5');
    expect(request.request.method).toBe('PUT');
    expect(request.request.body).toEqual(body);
    request.flush({ message: 'Temporada actualizada' });
  });

  it('deactivates a global season', () => {
    service.deactivate(5).subscribe();
    const request = http.expectOne('/cdv-api/seasons/5/deactivate');
    expect(request.request.method).toBe('PATCH');
    request.flush({ message: 'Temporada desactivada' });
  });
});
