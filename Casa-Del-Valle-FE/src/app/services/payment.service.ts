import { HttpClient } from "@angular/common/http";
import { RestConstants } from "../components/rest-constants";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { CreatePayment } from "../models/create-payment";
import { CreatePaymentResponse } from "../models/create-payment-response";
import { ReservationPayments } from "../models/reservation-payments";

@Injectable({
    providedIn: 'root',
})
export class PaymentService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    public getByReservation(reservationId: number): Observable<ReservationPayments> {
        return this.httpClient.get<ReservationPayments>(
            `${this.restConstants.getApiURL()}reservations/${reservationId}/payments`
        );
    }

    public create(reservationId: number, payment: CreatePayment): Observable<CreatePaymentResponse> {
        return this.httpClient.post<CreatePaymentResponse>(
            `${this.restConstants.getApiURL()}reservations/${reservationId}/payments`,
            payment
        );
    }
}
