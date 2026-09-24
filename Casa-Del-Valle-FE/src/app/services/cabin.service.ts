import { HttpClient } from "@angular/common/http";
import { HttpParams } from '@angular/common/http';
import { RestConstants } from "../components/rest-constants";
import { Injectable } from "@angular/core";
import { map, Observable } from "rxjs";
import { Cabin } from "../models/cabin";
import { CabinResponse, CabinsListResponse, CabinCardResponse, CabinCardsListResponse, CabinSearchParams } from "../models/cabin-response";
import { CabinRequest } from "../models/create-cabin";
import { RegisterResponse } from "../models/register-response";

@Injectable({
    providedIn: 'root'
})

export class CabinService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    public editCabin(cabinEdit: Cabin): Observable<{ message: string }> {
        if (!cabinEdit.id || cabinEdit.id <= 0) {
            throw new Error('ID de cabaña inválido para editar');
        }
        return this.httpClient.put<{ message: string }>(
            `${this.restConstants.getApiURL()}cabins/${cabinEdit.id}`, cabinEdit
        );
    }

    public deleteCabin(id: number): Observable<{ message: string; rows_affected?: number }> {
        return this.httpClient.delete<{ message: string; rows_affected?: number }>(
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

    public searchCabins(filters: CabinSearchParams): Observable<CabinCardsListResponse> {
        let params = new HttpParams();
        
        if (filters.name) params = params.set('name', filters.name);
        if (filters.host_name) params = params.set('host_name', filters.host_name);
        if (filters.host_id) params = params.set('host_id', filters.host_id.toString());
        if (filters.min_capacity) params = params.set('min_capacity', filters.min_capacity.toString());
        if (filters.max_capacity) params = params.set('max_capacity', filters.max_capacity.toString());
        if (filters.min_price) params = params.set('min_price', filters.min_price.toString());
        if (filters.max_price) params = params.set('max_price', filters.max_price.toString());

        return this.httpClient.get<CabinCardsListResponse>(
            `${this.restConstants.getApiURL()}cabins/search`, { params }
        );
    }
}
