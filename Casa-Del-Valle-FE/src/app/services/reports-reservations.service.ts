import { HttpClient, HttpParams } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { RestConstants } from "../components/rest-constants";
import { ReservationsReportFilter } from "../models/report-reservations";
import { ReservationsReportResponse } from "../models/report-reservations";

@Injectable({
    providedIn: 'root',
})

export class ReportReservationsService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    public getReservationsReport(filter: ReservationsReportFilter = {}): Observable<ReservationsReportResponse> {
        let params = new HttpParams();

        if (filter.host_id != null) {
            params = params.set('host_id', filter.host_id);
        }

        if (filter.cabin_id != null) {
            params = params.set('cabin_id', filter.cabin_id);
        }

        return this.httpClient.get<ReservationsReportResponse>(
            `${this.restConstants.getApiURL()}reports/reservations`, { params }
        );
    }
}