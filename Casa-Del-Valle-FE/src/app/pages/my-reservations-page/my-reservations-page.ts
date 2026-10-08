import { Component, OnInit, computed, signal } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { Reservation } from '../../models/reservation';
import { CancelReservationResponse } from '../../models/cancelReservationResponse';
import { ReservationService } from '../../services/reservation.service';
import { ReservationCard } from '../../components/reservation-card/reservation-card';

type Tab = 'upcoming' | 'finished' | 'cancelled';

@Component({
  imports: [ReservationCard],
  selector: 'app-my-reservations-page',
  styleUrl: './my-reservations-page.css',
  templateUrl: './my-reservations-page.html',
})
export class MyReservationsPage implements OnInit {
  userId: number | null = null;

  reservations = signal<Reservation[]>([]);
  tab = signal<Tab>('upcoming');
  isLoading = signal(false);
  loadError = signal(false);
  cancelMessage = signal<string | null>(null);

  visible = computed(() => {
    const list = this.reservations();

    switch (this.tab()) {
      case 'upcoming':
        return list.filter((r) => r.status === 'activa');
      case 'finished':
        return list.filter((r) => r.status === 'finalizada');
      case 'cancelled':
        return list.filter((r) => r.status === 'cancelada');
    }
  });

  emptyMessage = computed(() => {
    switch (this.tab()) {
      case 'upcoming':
        return 'No tienes reservaciones próximas.';
      case 'finished':
        return 'No tienes reservaciones finalizadas.';
      case 'cancelled':
        return 'No tienes reservaciones canceladas.';
    }
  });

  constructor(
    private auth: AuthService,
    private reservationService: ReservationService
  ) {}

  ngOnInit(): void {
    this.userId = this.auth.getCurrentUserId();
    this.getReservations();
  }

  getReservations(): void {
    if (!this.userId) return;

    this.loadError.set(false);
    this.isLoading.set(true);
    this.reservationService.getReservationsByUser(this.userId).subscribe({
      next: (list: Reservation[]) => {
        this.reservations.set(list);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Error al cargar reservaciones:', err);
        this.isLoading.set(false);
        this.loadError.set(true);
      },
    });
  }

  onCancelled(res: CancelReservationResponse): void {
    this.reservations.update((list) =>
      list.map((r) => (r.id === res.reservation.id ? { ...r, ...res.reservation } : r))
    );
    this.cancelMessage.set(res.message);
  }
}