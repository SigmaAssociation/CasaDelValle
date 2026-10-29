import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { RestConstants } from '../components/rest-constants';
import { AffectedReservation, CreateMaintenance, Maintenance, MaintenanceDecision, PendingMaintenance } from '../models/maintenance';

@Injectable({ providedIn: 'root' })
export class MaintenanceService {
  private readonly http = inject(HttpClient);
  private readonly api = new RestConstants().getApiURL();

  list(cabinId: number) {
    return this.http.get<Maintenance[]>(`${this.api}maintenance/cabin/${cabinId}`);
  }

  preview(cabinId: number, dates: { start_date: string; end_date: string }) {
    return this.http.post<{ reservations: AffectedReservation[] }>(
      `${this.api}maintenance/cabin/${cabinId}/preview`, dates,
    );
  }

  create(cabinId: number, request: CreateMaintenance) {
    return this.http.post<Maintenance>(`${this.api}maintenance/cabin/${cabinId}`, request);
  }

  pending(reservationId: number) {
    return this.http.get<PendingMaintenance>(`${this.api}maintenance/reservation/${reservationId}`);
  }

  resolve(reservationId: number, decision: MaintenanceDecision) {
    return this.http.post<{message: string}>(`${this.api}maintenance/reservation/${reservationId}/resolve`, decision);
  }

  cancel(cabinId: number, id: number) {
    return this.http.patch<Maintenance>(`${this.api}maintenance/cabin/${cabinId}/${id}/cancel`, {});
  }
}
