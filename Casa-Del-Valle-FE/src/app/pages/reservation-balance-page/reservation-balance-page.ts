import { CurrencyPipe, DatePipe } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { ReservationBalance } from '../../models/reservation-balance';
import { ReservationService } from '../../services/reservation.service';

type LoadError = 'forbidden' | 'not-found' | 'generic';

const STATUS_LABELS: Record<string, string> = {
  activa: 'Activa',
  finalizada: 'Finalizada',
  cancelada: 'Cancelada',
};

const PAYMENT_BADGES: Record<string, { label: string; classes: string }> = {
  pendiente: { label: 'Pago pendiente', classes: 'bg-orange text-white' },
  pagado: { label: 'Saldada', classes: 'bg-green text-white' },
  saldo_a_favor: { label: 'Saldo a favor', classes: 'bg-brown text-white' },
};

@Component({
  imports: [CurrencyPipe, DatePipe, RouterLink],
  selector: 'app-reservation-balance-page',
  templateUrl: './reservation-balance-page.html',
})
export class ReservationBalancePage {
  private route = inject(ActivatedRoute);
  private reservationService = inject(ReservationService);

  reservationId = signal<number | null>(null);
  balance = signal<ReservationBalance | null>(null);
  isLoading = signal(false);
  error = signal<LoadError | null>(null);

  statusLabel = computed(() => {
    const status = this.balance()?.status ?? '';
    return STATUS_LABELS[status] ?? status;
  });

  paymentBadge = computed(() => {
    const status = this.balance()?.paymentStatus ?? 'pendiente';
    return PAYMENT_BADGES[status] ?? PAYMENT_BADGES['pendiente'];
  });

  /** Texto bajo el monto principal, según el estado del saldo. */
  balanceHint = computed(() => {
    switch (this.balance()?.paymentStatus) {
      case 'pagado':
        return 'La reservación está completamente pagada.';
      case 'saldo_a_favor':
        return 'Se ha pagado más del total; hay un saldo a favor del huésped.';
      default:
        return 'Monto que aún falta por pagar.';
    }
  });

  balanceLabel = computed(() =>
    this.balance()?.paymentStatus === 'saldo_a_favor' ? 'Saldo a favor' : 'Saldo pendiente'
  );

  /** El saldo a favor se muestra en positivo; el estado ya indica su sentido. */
  balanceAmount = computed(() => Math.abs(this.balance()?.pendingBalance ?? 0));

  errorMessage = computed(() => {
    switch (this.error()) {
      case 'forbidden':
        return 'No tienes permisos para ver el saldo de esta reservación.';
      case 'not-found':
        return 'No se encontró la reservación.';
      case 'generic':
        return 'No se pudo cargar el saldo de la reservación.';
      default:
        return '';
    }
  });

  constructor() {
    this.route.paramMap.pipe(takeUntilDestroyed()).subscribe((params) => {
      const id = Number(params.get('id'));
      if (!Number.isInteger(id) || id <= 0) {
        this.reservationId.set(null);
        this.balance.set(null);
        this.error.set('not-found');
        return;
      }
      this.reservationId.set(id);
      this.load();
    });
  }

  load(): void {
    const id = this.reservationId();
    if (!id) return;

    this.error.set(null);
    this.isLoading.set(true);
    this.reservationService.getBalance(id).subscribe({
      next: (balance) => {
        this.balance.set(balance);
        this.isLoading.set(false);
      },
      error: (err: unknown) => {
        console.error('Error al cargar el saldo de la reservación:', err);
        this.balance.set(null);
        this.isLoading.set(false);
        const status = err instanceof HttpErrorResponse ? err.status : 0;
        this.error.set(status === 403 ? 'forbidden' : status === 404 ? 'not-found' : 'generic');
      },
    });
  }
}