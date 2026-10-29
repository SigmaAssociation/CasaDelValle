import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { PriceQuote } from '../models/price-quote';
import { RestConstants } from '../components/rest-constants';
import { CabinSeasonPrice, CabinSeasonPriceRequest, Season } from '../models/season';

@Injectable({ providedIn: 'root' })
export class SeasonService {
  private readonly api = new RestConstants().getApiURL();

  constructor(private http: HttpClient) {}

  quote(cabinId: number, startDate: string, endDate: string): Observable<PriceQuote> {
    return this.http.post<PriceQuote>(`${this.api}seasons/quote`, { cabinId, startDate, endDate });
  }

  list(): Observable<Season[]> {
    return this.http.get<Season[]>(`${this.api}seasons`);
  }

  listCabinPrices(cabinId: number): Observable<CabinSeasonPrice[]> {
    return this.http.get<CabinSeasonPrice[]>(`${this.api}cabin-season-prices/cabin/${cabinId}`);
  }

  createCabinPrice(cabinId: number, request: CabinSeasonPriceRequest): Observable<CabinSeasonPrice> {
    return this.http.post<CabinSeasonPrice>(`${this.api}cabin-season-prices/cabin/${cabinId}`, request);
  }

  deleteCabinPrice(id: number): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.api}cabin-season-prices/${id}`);
  }
}
