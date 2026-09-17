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

    public editCabin(cabinEdit: Cabin): Observable<{ mensaje: string }> {
        if (!cabinEdit.id || cabinEdit.id <= 0) {
            throw new Error('ID de cabaña inválido para editar');
        }
        return this.httpClient.put<{ mensaje: string }>(
            `${this.restConstants.getApiURL()}cabins/${cabinEdit.id}`, cabinEdit
        );
    }

    public deleteCabin(id: number): Observable<{ mensaje: string; rows_affected?: number }> {
        return this.httpClient.delete<{ mensaje: string; rows_affected?: number }>(
            `${this.restConstants.getApiURL()}cabins/${id}`
        );
    }

    public createCabin(cabin: CabinRequest): Observable<RegisterResponse> {
        return this.httpClient.post<RegisterResponse>(
            `${this.restConstants.getApiURL()}cabins`, cabin
        );
    }

    public getCabins(): Observable<Cabin[]> {
        return this.httpClient.get<CabinsListResponse | Cabin[]>(
            `${this.restConstants.getApiURL()}cabins`
        ).pipe(map((res) => Array.isArray(res) ? res : res.data));
    }

    public getCabinsByUserId(userId: number): Observable<Cabin[]> {
        return this.httpClient.get<CabinsListResponse | Cabin[]>(
            `${this.restConstants.getApiURL()}cabins/user/${userId}`
        ).pipe(map((res) => Array.isArray(res) ? res : res.data));
    }

    public getCabinById(id: number): Observable<Cabin> {
        return this.httpClient.get<CabinResponse>(
            `${this.restConstants.getApiURL()}cabins/${id}`
        );
    }

    // DELETE /cdv-api/cabins/{id} — eliminar cabaña.
    public deleteCabin(id: number): Observable<void> {
        return this.httpClient.delete<void>(
            `${this.restConstants.getApiURL()}cabins/${id}`
        );
    }
}
