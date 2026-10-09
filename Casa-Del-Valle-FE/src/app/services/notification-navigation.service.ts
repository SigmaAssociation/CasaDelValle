import { Injectable } from "@angular/core";
import { Observable, catchError, map, of } from "rxjs";
import { Notification } from "../models/notification";
import { ReservationService } from "./reservation.service";

@Injectable({
    providedIn: 'root',
})
export class NotificationNavigationService {

    constructor(private reservationService: ReservationService) { }

    // Destino inmediato cuando la notificación ya trae el ID de la vista.
    // Las reservaciones se resuelven aparte porque el aviso trae el ID de
    // la reservación y la vista es la de la cabaña (ver resolveLink).
    public linkFor(notification: Notification): string {
        switch (notification.entidad_tipo) {
            case 'cabana':
                return notification.entidad_id ? `/cabins/${notification.entidad_id}` : '/mis-cabanas';
            case 'reservacion':
                return '/mis-reservaciones';
            case 'usuario':
                return '/perfil';
            default:
                return '/notificaciones';
        }
    }

    // Destino final al hacer click: las reservaciones llevan a la vista
    // detallada de su cabaña; si no se puede resolver, al fallback por tipo.
    public resolveLink(notification: Notification): Observable<string> {
        if (notification.entidad_tipo === 'reservacion' && notification.entidad_id) {
            return this.reservationService.getById(notification.entidad_id).pipe(
                map((detail) => `/cabins/${detail.cabin_id}`),
                catchError(() => of(this.linkFor(notification))),
            );
        }
        return of(this.linkFor(notification));
    }
}
