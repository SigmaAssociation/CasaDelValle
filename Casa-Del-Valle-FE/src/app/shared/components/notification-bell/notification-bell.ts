import {
  Component,
  ElementRef,
  HostListener,
  OnInit,
  inject,
  signal,
} from '@angular/core';
import { NavigationEnd, Router, RouterModule } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { filter, interval } from 'rxjs';
import { NotificationService } from '../../../services/notification.service';
import { Notification } from '../../../models/notification';

type NotificationCategory = 'reserva' | 'cabana' | 'cuenta';

@Component({
  selector: 'app-notification-bell',
  imports: [RouterModule],
  templateUrl: './notification-bell.html',
  styleUrl: './notification-bell.css',
})
export class NotificationBell implements OnInit {
  private readonly PREVIEW_LIMIT = 8;
  private readonly POLLING_MS = 60_000;

  private notificationService = inject(NotificationService);
  private router = inject(Router);
  private elementRef = inject(ElementRef<HTMLElement>);

  readonly isOpen = signal(false);
  readonly isLoading = signal(false);
  readonly loadError = signal(false);
  readonly unreadCount = signal(0);
  readonly items = signal<Notification[]>([]);

  constructor() {
    this.router.events
      .pipe(
        filter((event): event is NavigationEnd => event instanceof NavigationEnd),
        takeUntilDestroyed()
      )
      .subscribe(() => {
        if (!this.isOpen()) {
          this.refreshUnreadCount();
        }
      });

    // Refresco periódico: mantiene el contador y el panel al día sin recargar
    // la página. El refresco del panel es silencioso para no parpadear.
    interval(this.POLLING_MS)
      .pipe(takeUntilDestroyed())
      .subscribe(() => {
        this.refreshUnreadCount();
        if (this.isOpen()) {
          this.loadNotifications(true);
        }
      });
  }

  ngOnInit(): void {
    this.refreshUnreadCount();
  }

  toggle(): void {
    if (this.isOpen()) {
      this.close();
      return;
    }
    this.isOpen.set(true);
    this.loadNotifications();
  }

  close(): void {
    this.isOpen.set(false);
  }

  refreshUnreadCount(): void {
    this.notificationService.getUnreadCount().subscribe({
      next: (count) => this.unreadCount.set(count),
      error: (err) => console.error('Error al cargar el contador de notificaciones:', err),
    });
  }

  loadNotifications(silent = false): void {
    this.loadError.set(false);
    if (!silent) {
      this.isLoading.set(true);
    }
    this.notificationService.getNotifications(this.PREVIEW_LIMIT, 0).subscribe({
      next: (res) => {
        this.items.set(res.data.slice(0, this.PREVIEW_LIMIT));
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Error al cargar las notificaciones:', err);
        this.isLoading.set(false);
        if (!silent) {
          this.loadError.set(true);
        }
      },
    });
  }

  markAllRead(): void {
    this.notificationService.markAllRead().subscribe({
      next: () => {
        this.unreadCount.set(0);
        this.items.update((list) => list.map((n) => ({ ...n, leida: true })));
      },
      error: (err) => console.error('Error al marcar las notificaciones como leídas:', err),
    });
  }

  openItem(notification: Notification): void {
    if (!notification.leida) {
      this.notificationService.markRead(notification.id).subscribe({
        next: () => {
          this.items.update((list) =>
            list.map((n) => (n.id === notification.id ? { ...n, leida: true } : n))
          );
          this.unreadCount.update((count) => Math.max(0, count - 1));
        },
        error: (err) => console.error('Error al marcar la notificación como leída:', err),
      });
    }
    this.close();
    this.router.navigateByUrl(this.linkFor(notification));
  }

  // linkFor resuelve a dónde llevar al usuario según la entidad de la
  // notificación; si no hay una vista relacionada, va al buzón completo.
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
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: MouseEvent): void {
    if (!this.isOpen()) {
      return;
    }
    const target = event.target as Node | null;
    if (target && !this.elementRef.nativeElement.contains(target)) {
      this.close();
    }
  }

  @HostListener('document:keydown.escape')
  onEscape(): void {
    this.close();
  }
}
