export interface Notification {
  id: number;
  user_id: number;
  tipo: string;
  mensaje: string;
  entidad_tipo?: string | null;
  entidad_id?: number | null;
  leida: boolean;
  fecha_creacion: string;
}

export interface NotificationListResponse {
  data: Notification[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
}

export interface UnreadCountResponse {
  unread_count: number;
}

export interface MarkAllReadResponse {
  message: string;
  updated: number;
}

export interface MarkReadResponse {
  message: string;
}
