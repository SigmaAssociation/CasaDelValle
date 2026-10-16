import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { RestConstants } from '../components/rest-constants';
import {
  CreateRefundRequest,
  CreateRefundResponse,
  RefundsHistoryResponse,
} from '../models/refund';

@Injectable({
  providedIn: 'root',
})
export class RefundService {
  private restConstants = new RestConstants();

  constructor(private httpClient: HttpClient) {}

  public getHistory(reservationId: number): Observable<RefundsHistoryResponse> {
    return this.httpClient.get<RefundsHistoryResponse>(
      `${this.restConstants.getApiURL()}refunds/reservation/${reservationId}`
    );
  }

  public create(
    reservationId: number,
    request: CreateRefundRequest = {}
  ): Observable<CreateRefundResponse> {
    return this.httpClient.post<CreateRefundResponse>(
      `${this.restConstants.getApiURL()}refunds/reservation/${reservationId}`,
      request
    );
  }
}
