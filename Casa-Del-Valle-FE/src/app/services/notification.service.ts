import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { map, Observable } from 'rxjs';
import { RestConstants } from '../components/rest-constants';
import {
  MarkAllReadResponse,
  MarkReadResponse,
  NotificationListResponse,
  UnreadCountResponse,
} from '../models/notification';

@Injectable({
  providedIn: 'root',
})
export class NotificationService {
  private restConstants = new RestConstants();

  constructor(private httpClient: HttpClient) {}

  public getNotifications(limit?: number, offset?: number): Observable<NotificationListResponse> {
    let params = new HttpParams();
    if (limit != null) {
      params = params.set('limit', limit);
    }
    if (offset != null) {
      params = params.set('offset', offset);
    }

    return this.httpClient
      .get<NotificationListResponse>(`${this.restConstants.getApiURL()}notifications`, { params })
      .pipe(map((res) => ({ ...res, data: res?.data ?? [] })));
  }

  public getUnreadCount(): Observable<number> {
    return this.httpClient
      .get<UnreadCountResponse>(`${this.restConstants.getApiURL()}notifications/unread-count`)
      .pipe(map((res) => res?.unread_count ?? 0));
  }

  public markAllRead(): Observable<MarkAllReadResponse> {
    return this.httpClient.post<MarkAllReadResponse>(
      `${this.restConstants.getApiURL()}notifications/read-all`,
      {}
    );
  }

  public markRead(id: number): Observable<MarkReadResponse> {
    return this.httpClient.patch<MarkReadResponse>(
      `${this.restConstants.getApiURL()}notifications/${id}/read`,
      {}
    );
  }
}
