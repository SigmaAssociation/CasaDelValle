import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { RestConstants } from "../components/rest-constants";
import { CabinImage, CabinImagesListResponse } from "../models/cabin-image";

@Injectable({
    providedIn: 'root'
})

export class CabinImageService {
    restConstants = new RestConstants();

    constructor(private httpClient: HttpClient) { }

    // Todas las imágenes de una cabaña, de la más reciente a la más antigua.
    public getImagesByCabinId(cabinId: number): Observable<CabinImagesListResponse> {
        return this.httpClient.get<CabinImagesListResponse>(
            `${this.restConstants.getApiURL()}cabins/images/${cabinId}`
        );
    }

    // Imagen principal de una cabaña: la más reciente.
    public getMainImageByCabinId(cabinId: number): Observable<CabinImage> {
        return this.httpClient.get<CabinImage>(
            `${this.restConstants.getApiURL()}cabins/images/${cabinId}/principal`
        );
    }

    // Una imagen por su identificador.
    public getImageById(imageId: number): Observable<CabinImage> {
        return this.httpClient.get<CabinImage>(
            `${this.restConstants.getApiURL()}images/${imageId}`
        );
    }
}