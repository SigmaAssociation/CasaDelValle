import { HttpClient } from "@angular/common/http";
import { RestConstants } from "../shared/rest-constants";
import { UserRequest } from "../models/create-user";
import { UpdateUserRequest } from "../models/update-user";
import { User } from "../models/user";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { RegisterResponse } from "../models/register-response";
import { UpdateUserResponse } from "../models/update-user-response";

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

    public getUserById(id: number): Observable<User> {
        return this.httpClient.get<User>(
            `${this.restConstants.getApiURL()}users/${id}`
        );
    }

    public updateUser(id: number, user: UpdateUserRequest): Observable<UpdateUserResponse> {
        return this.httpClient.put<UpdateUserResponse>(
            `${this.restConstants.getApiURL()}users/${id}`, user
        );
    }
}