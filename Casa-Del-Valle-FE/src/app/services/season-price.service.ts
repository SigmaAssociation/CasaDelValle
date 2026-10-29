import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { RestConstants } from "../components/rest-constants";
import { DeleteSeasonPriceResponse } from "../models/sesion-price";

@Injectable({
    providedIn: 'root',
})

export class SeasonPriceService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    public delete(cabinId: number, id: number): Observable<DeleteSeasonPriceResponse> {
        return this.httpClient.delete<DeleteSeasonPriceResponse>(
            `${this.restConstants.getApiURL()}cabins/${cabinId}/season-prices/${id}`
        );
    }
}