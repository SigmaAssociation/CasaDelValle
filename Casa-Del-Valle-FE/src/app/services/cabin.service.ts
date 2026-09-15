import { HttpClient } from "@angular/common/http";
import { RestConstants } from "../components/rest-constants";
import { Injectable } from "@angular/core";
import { map, Observable } from "rxjs";
import { Cabin } from "../models/cabin";
import { CabinResponse, CabinsListResponse } from "../models/cabin-response";
import { CabinRequest } from "../models/create-cabin";
import { RegisterResponse } from "../models/register-response";

@Injectable({
    providedIn: 'root'
})

export class CabinService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    public editCabin(cabinEdit: Cabin): Observable<void> {
        return this.httpClient.put<void>(
            `${this.restConstants.getApiURL()}cabins`, cabinEdit
        );
    }

    public createCabin(cabin: CabinRequest): Observable<RegisterResponse> {
        return this.httpClient.post<RegisterResponse>(
            `${this.restConstants.getApiURL()}cabins`, cabin
        );
    }

    // GET /cdv-api/cabins — todas las cabañas.
    // Desenvuelve el struct CabinsListResponse { data, total } del BE.
    public getCabins(): Observable<Cabin[]> {
        return this.httpClient.get<CabinsListResponse | Cabin[]>(
            `${this.restConstants.getApiURL()}cabins`
        ).pipe(map((res) => Array.isArray(res) ? res : res.data));
    }

    // GET /cdv-api/cabins/user/{userId} — cabañas por usuario (anfitrión).
    public getCabinsByUserId(userId: number): Observable<Cabin[]> {
        return this.httpClient.get<CabinsListResponse | Cabin[]>(
            `${this.restConstants.getApiURL()}cabins/user/${userId}`
        ).pipe(map((res) => Array.isArray(res) ? res : res.data));
    }

    // GET /cdv-api/cabins/{id} — cabaña por id.
    public getCabinById(id: number): Observable<Cabin> {
        return this.httpClient.get<CabinResponse>(
            `${this.restConstants.getApiURL()}cabins/${id}`
        );
    }
}
