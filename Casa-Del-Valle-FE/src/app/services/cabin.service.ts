import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { CabinRequest } from '../models/create-cabin';
import { RegisterResponse } from '../models/register-response';
import { RestConstants } from '../components/rest-constants';

@Injectable({
  providedIn: 'root'
})
export class CabinService {
  restConstants = new RestConstants();

  constructor(private httpClient: HttpClient) {}

  public createCabin(cabin: CabinRequest): Observable<RegisterResponse> {
    return this.httpClient.post<RegisterResponse>(
      `${this.restConstants.getApiURL()}cabins`,
      cabin
    );
  }
}
