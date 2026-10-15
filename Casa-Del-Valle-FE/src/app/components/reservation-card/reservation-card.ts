import { Component, EventEmitter, inject, Input, Output, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Reservation } from '../../models/reservation';
import { ReservationService } from '../../services/reservation.service';
import { CancelReservationResponse } from '../../models/cancelReservationResponse';
import { CheckinReservationResponse } from '../../models/checkin-reservation-response';
import { CancelReservationModal } from '../cancel-reservation-modal/cancel-reservation-modal';

@Component({
  selector: 'app-reservation-card',
  imports: [CancelReservationModal, RouterLink],
  templateUrl: './reservation-card.html',
})
export class ReservationCard {
  private reservationService = inject(ReservationService);

  @Input({ required: true }) reservation!: Reservation;
  // Muestra la acción de check-in (solo tiene sentido para el anfitrión).
  @Input() showCheckin = false;
  @Output() cancelled = new EventEmitter<CancelReservationResponse>();
  @Output() checkedIn = new EventEmitter<CheckinReservationResponse>();

  showModal = signal(false);
  isCancelling = signal(false);
  isCheckingIn = signal(false);
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

  get isCancelled(): boolean {
    return this.reservation.status === 'cancelada';
  }

  get isActive(): boolean {
    return this.reservation.status === 'activa';
  }

  registerCheckin(): void {
    if (!this.showCheckin || !this.isActive || this.isCheckingIn()) {
      return;
    }
    if (!confirm(`¿Registrar el check-in de la reservación #${this.reservation.id}? Se valida que el pago esté completo y que esté dentro de las fechas.`)) {
      return;
    }
    this.isCheckingIn.set(true);
    this.errorMessage.set(null);
    this.reservationService.checkin(this.reservation.id).subscribe({
      next: (res: CheckinReservationResponse) => {
        this.isCheckingIn.set(false);
        this.checkedIn.emit(res);
      },
      error: (err) => {
        this.isCheckingIn.set(false);
        this.errorMessage.set(
          err.error?.message ?? err.error?.error ?? 'No se pudo registrar el check-in. Intenta de nuevo.'
        );
      },
    });
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
    if (this.isCancelled) return false;
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