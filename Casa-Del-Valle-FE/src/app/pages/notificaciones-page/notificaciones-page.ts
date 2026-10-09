import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { Notification } from '../../models/notification';
import { NotificationService } from '../../services/notification.service';

type Filter = 'todas' | 'no_leidas';
type NotificationCategory = 'reserva' | 'cabana' | 'cuenta';

@Component({
  selector: 'app-notificaciones-page',
  imports: [],
  templateUrl: './notificaciones-page.html',
  styleUrl: './notificaciones-page.css',
})
export class NotificacionesPage implements OnInit {
  private readonly PAGE_SIZE = 20;

  private notificationService = inject(NotificationService);
  private router = inject(Router);

  readonly notifications = signal<Notification[]>([]);
  readonly isLoading = signal(false);
  readonly isLoadingMore = signal(false);
  readonly loadError = signal(false);
  readonly filter = signal<Filter>('todas');
  readonly actionMessage = signal<string | null>(null);
  readonly unreadCount = signal(0);
  readonly total = signal(0);
  readonly hasMore = signal(false);

  readonly sorted = computed(() =>
    [...this.notifications()].sort((a, b) => {
      const diff = new Date(b.fecha_creacion).getTime() - new Date(a.fecha_creacion).getTime();
      return diff !== 0 ? diff : b.id - a.id;
    })
  );

  readonly visible = computed(() => {
    const list = this.sorted();
    return this.filter() === 'no_leidas' ? list.filter((n) => !n.leida) : list;
  });

  readonly emptyMessage = computed(() =>
    this.filter() === 'no_leidas'
      ? 'No tienes notificaciones sin leer.'
      : 'Aún no tienes notificaciones.'
  );

  ngOnInit(): void {
    this.loadNotifications();
  }

  loadNotifications(): void {
    this.loadError.set(false);
    this.isLoading.set(true);
    this.notificationService.getNotifications(this.PAGE_SIZE, 0).subscribe({
      next: (res) => {
        this.notifications.set(res.data);
        this.total.set(res.total);
        this.hasMore.set(res.has_more);
        this.isLoading.set(false);
        this.refreshUnreadCount();
      },
      error: (err) => {
        console.error('Error al cargar las notificaciones:', err);
        this.isLoading.set(false);
        this.loadError.set(true);
      },
    });
  }

  loadMore(): void {
    if (this.isLoadingMore() || !this.hasMore()) {
      return;
    }
    this.isLoadingMore.set(true);
    const offset = this.notifications().length;
    this.notificationService.getNotifications(this.PAGE_SIZE, offset).subscribe({
      next: (res) => {
        this.notifications.update((list) => [...list, ...res.data]);
        this.total.set(res.total);
        this.hasMore.set(res.has_more);
        this.isLoadingMore.set(false);
      },
      error: (err) => {
        console.error('Error al cargar más notificaciones:', err);
        this.isLoadingMore.set(false);
      },
    });
  }

  refreshUnreadCount(): void {
    this.notificationService.getUnreadCount().subscribe({
      next: (count) => this.unreadCount.set(count),
      error: (err) => console.error('Error al cargar el contador de notificaciones:', err),
    });
  }

  openItem(notification: Notification): void {
    this.markRead(notification);
    this.router.navigateByUrl(this.linkFor(notification));
  }

  // linkFor resuelve a dónde llevar al usuario según la entidad de la
  // notificación; si no hay una vista relacionada, permanece en el buzón.
  linkFor(notification: Notification): string {
    switch (notification.entidad_tipo) {
      case 'reservacion':
        return '/mis-reservaciones';
      case 'cabana':
        return '/mis-cabanas';
      case 'usuario':
        return '/perfil';
      default:
        return '/notificaciones';
    }
  }

  markAllRead(): void {
    if (this.unreadCount() === 0) {
      return;
    }
    this.notificationService.markAllRead().subscribe({
      next: (res) => {
        this.notifications.update((list) => list.map((n) => ({ ...n, leida: true })));
        this.unreadCount.set(0);
        this.actionMessage.set(res.message ?? 'Notificaciones marcadas como leídas.');
      },
      error: (err) => console.error('Error al marcar las notificaciones como leídas:', err),
    });
  }

  markRead(notification: Notification): void {
    if (notification.leida) {
      return;
    }
    this.notificationService.markRead(notification.id).subscribe({
      next: () => {
        this.notifications.update((list) =>
          list.map((n) => (n.id === notification.id ? { ...n, leida: true } : n))
        );
        this.unreadCount.update((count) => Math.max(0, count - 1));
      },
      error: (err) => console.error('Error al marcar la notificación como leída:', err),
    });
  }

  category(notification: Notification): NotificationCategory {
    if (notification.tipo.startsWith('reserva')) {
      return 'reserva';
    }
    if (notification.tipo.startsWith('cabana')) {
      return 'cabana';
    }
    return 'cuenta';
  }

  formatDate(value: string): string {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return '';
    }
    return date.toLocaleDateString('es-GT', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }
}
