import { HttpClient } from "@angular/common/http";
import { RestConstants } from "../components/rest-constants";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { Reservation } from "../models/reservation";
import { ReservationDetail } from "../models/reservation-detail";
import { CheckinReservationResponse } from "../models/checkin-reservation-response";
import { CancelReservationResponse } from "../models/cancelReservationResponse";
import { PayAnticipoResponse } from "../models/pay-anticipo-response";
import { ReservationRequest } from "../models/create-reservation";
import { CreateReservationResponse } from "../models/create-reservation-response";

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

    public getReservationsByCabin(cabinId: number): Observable<Reservation[]> {
        return this.httpClient.get<Reservation[]>(
            `${this.restConstants.getApiURL()}reservations/cabin/${cabinId}`
        );
    }

    public getById(id: number): Observable<ReservationDetail> {
        return this.httpClient.get<ReservationDetail>(
            `${this.restConstants.getApiURL()}reservations/${id}`
        );
    }

    public create(reservation: ReservationRequest): Observable<CreateReservationResponse> {
        return this.httpClient.post<CreateReservationResponse>(
            `${this.restConstants.getApiURL()}reservations`,
            reservation
        );
    }

    public cancel(id: number): Observable<CancelReservationResponse> {
        return this.httpClient.patch<CancelReservationResponse>(
            `${this.restConstants.getApiURL()}reservations/${id}/cancel`,
            {}
        );
    }

    public checkin(id: number, cashAmount = 0): Observable<CheckinReservationResponse> {
        return this.httpClient.patch<CheckinReservationResponse>(
            `${this.restConstants.getApiURL()}reservations/${id}/checkin`,
            { cash_amount: cashAmount }
        );
    }

    public payAnticipo(id: number, payload?: { amount?: number; method?: string }): Observable<PayAnticipoResponse> {
        return this.httpClient.post<PayAnticipoResponse>(
            `${this.restConstants.getApiURL()}reservations/${id}/payment`,
            payload ?? {}
        );
    }
}