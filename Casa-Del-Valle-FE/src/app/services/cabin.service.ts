import { HttpClient } from "@angular/common/http";
import { RestConstants } from "../shared/rest-constants";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { Cabin } from "../models/cabin";

@Injectable({
    providedIn: 'root'
})

export class CabinService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    public editCabin(cabinEdit: Cabin) {
        return this.httpClient.put<void>(
            `${this.restConstants.getApiURL()}cabins`, cabinEdit
        );
    }

    public getCabinsByUserId(userId: number) {
        return this.httpClient.get<Cabin[]>(
            `${this.restConstants.getApiURL()}cabins/user/${userId}`
        );
    }

    public getCabinById(id: number): Observable<Cabin> {
        return this.httpClient.get<Cabin>(
            `${this.restConstants.getApiURL()}cabins/${id}`
        );
    }


}
