import { HttpClient } from "@angular/common/http";
import { RestConstants } from "../components/rest-constants";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { Reservation } from "../models/reservation";
import { CancelReservationResponse } from "../models/cancelReservationResponse";

@Injectable({
    providedIn: 'root',
})

export class ReservationService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    public getReservationsByUser(id: number): Observable<Reservation[]> {
        return this.httpClient.get<Reservation[]>(
            `${this.restConstants.getApiURL()}reservations/user/${id}`
        );
    }

    public cancel(id: number): Observable<CancelReservationResponse> {
        return this.httpClient.patch<CancelReservationResponse>(
            `${this.restConstants.getApiURL()}reservations/${id}/cancel`,
            {}
        );
    }

}