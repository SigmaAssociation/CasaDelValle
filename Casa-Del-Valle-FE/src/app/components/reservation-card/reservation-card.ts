import { Component, EventEmitter, inject, Input, Output, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Reservation } from '../../models/reservation';
import { ReservationService } from '../../services/reservation.service';
import { AuthService } from '../../services/auth.service';
import { CancelReservationResponse } from '../../models/cancelReservationResponse';
import { PayAnticipoResponse } from '../../models/pay-anticipo-response';
import { CancelReservationModal } from '../cancel-reservation-modal/cancel-reservation-modal';
import { PayAnticipoModal } from '../pay-anticipo-modal/pay-anticipo-modal';

@Component({
  selector: 'app-reservation-card',
  imports: [CancelReservationModal, PayAnticipoModal, RouterLink],
  templateUrl: './reservation-card.html',
})
export class ReservationCard {
  private reservationService = inject(ReservationService);
  private auth = inject(AuthService);

  @Input({ required: true }) reservation!: Reservation;
  @Output() cancelled = new EventEmitter<CancelReservationResponse>();
  @Output() paid = new EventEmitter<PayAnticipoResponse>();

  showModal = signal(false);
  isCancelling = signal(false);
  showPayModal = signal(false);
  isPaying = signal(false);
  errorMessage = signal<string | null>(null);

  openModal(): void {
    if (!this.canCancel) return;
    this.errorMessage.set(null);
    this.showModal.set(true);
  }

  closeModal(): void {
    if (this.isCancelling()) return;
    this.showModal.set(false);
  }

  confirmCancel(): void {
    this.isCancelling.set(true);
    this.reservationService.cancel(this.reservation.id).subscribe({
      next: (res: CancelReservationResponse) => {
        this.isCancelling.set(false);
        this.showModal.set(false);
        this.cancelled.emit(res);
      },
      error: (err) => {
        this.isCancelling.set(false);
        this.showModal.set(false);
        this.errorMessage.set(
          err.error?.message ?? err.error?.error ?? 'No se pudo cancelar la reservación. Intenta de nuevo.'
        );
      },
    });
  }

  openPayModal(): void {
    if (!this.canPay) return;
    this.errorMessage.set(null);
    this.showPayModal.set(true);
  }

  closePayModal(): void {
    if (this.isPaying()) return;
    this.showPayModal.set(false);
  }

  confirmPay(): void {
    this.isPaying.set(true);
    this.reservationService.payAnticipo(this.reservation.id).subscribe({
      next: (res: PayAnticipoResponse) => {
        this.isPaying.set(false);
        this.showPayModal.set(false);
        this.paid.emit(res);
      },
      error: (err) => {
        this.isPaying.set(false);
        this.showPayModal.set(false);
        this.errorMessage.set(
          err.error?.message ?? err.error?.error ?? 'No se pudo registrar el anticipo. Intenta de nuevo.'
        );
      },
    });
  }

  get isCancelled(): boolean {
    return this.reservation.status === 'cancelada';
  }

  get isPending(): boolean {
    return this.reservation.status === 'pendiente';
  }

  get isConfirmed(): boolean {
    return this.reservation.status === 'confirmada';
  }

  get isInProgress(): boolean {
    return this.reservation.status === 'en_curso';
  }

  get isFinished(): boolean {
    return this.reservation.status === 'finalizada';
  }

  // Solo quien reservó (el huésped) puede pagar el anticipo.
  get isOwnReservation(): boolean {
    const userId = this.auth.getCurrentUserId();
    return userId != null && userId === this.reservation.guestId;
  }

  get canPay(): boolean {
    return this.isPending && this.isOwnReservation;
  }

  get anticipoAmount(): number {
    return Math.round(this.reservation.totalPrice * 0.3 * 100) / 100;
  }

  get nights(): number {
    const ms =
      this.parse(this.reservation.endDate).getTime() -
      this.parse(this.reservation.startDate).getTime();
    return Math.round(ms / 86_400_000);
  }

  get cancelDeadline(): Date {
    const d = this.parse(this.reservation.startDate);
    d.setDate(d.getDate() - 3);
    return d;
  }

  get canCancel(): boolean {
    if (!this.isPending && !this.isConfirmed) return false;
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return today <= this.cancelDeadline;
  }

  formatDate(value: string): string {
    return this.parse(value).toLocaleDateString('es-GT', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    });
  }

  formatPrice(value: number): string {
    return `Q ${value.toLocaleString('es-GT', { minimumFractionDigits: 2 })}`;
  }

  formatDeadline(): string {
    return this.cancelDeadline.toLocaleDateString('es-GT', {
      day: 'numeric',
      month: 'long',
    });
  }

  private parse(value: string): Date {
    return new Date(value.length === 10 ? `${value}T00:00:00` : value);
  }
}
