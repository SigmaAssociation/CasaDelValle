import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { RestConstants } from '../components/rest-constants';
import { CabinAvailability } from '../models/cabin-availability';

@Injectable({
  providedIn: 'root'
})
export class CabinAvailabilityService {
  private readonly http = inject(HttpClient);
  private readonly restConstants = new RestConstants();

  getAvailability(cabinId: number): Observable<CabinAvailability> {
    return this.http.get<CabinAvailability>(`${this.restConstants.getApiURL()}cabins/availability/${cabinId}`);
  }
}
