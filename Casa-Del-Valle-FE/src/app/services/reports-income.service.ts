import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { RestConstants } from '../components/rest-constants';
import { IncomeReportFilter, IncomeReportResponse } from '../models/report-income';

@Injectable({
  providedIn: 'root',
})
export class ReportIncomeService {
  private readonly restConstants = new RestConstants();

  constructor(private httpClient: HttpClient) {}

  public getIncomeReport(filter: IncomeReportFilter): Observable<IncomeReportResponse> {
    let params = new HttpParams()
      .set('start_date', filter.startDate)
      .set('end_date', filter.endDate);

    if (filter.cabinId != null) {
      params = params.set('cabin_id', filter.cabinId);
    }

    return this.httpClient.get<IncomeReportResponse>(
      `${this.restConstants.getApiURL()}reports/income`,
      { params },
    );
  }
}
