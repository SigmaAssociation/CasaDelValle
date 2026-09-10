import { HttpClient } from "@angular/common/http";
import { RestConstants } from "../components/rest-constants";
import { UserRequest } from "../models/create-user";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { RegisterResponse } from "../models/register-response";

@Injectable({
    providedIn: 'root',
})

export class UserService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    public createUser(user: UserRequest): Observable<RegisterResponse> {
        return this.httpClient.post<RegisterResponse>(
            `${this.restConstants.getApiURL()}users`, user
        );
    }
}