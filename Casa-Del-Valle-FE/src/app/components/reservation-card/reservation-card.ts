import { Component, EventEmitter, inject, Input, Output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { Reservation } from '../../models/reservation';
import { ReservationService } from '../../services/reservation.service';
import { AuthService } from '../../services/auth.service';
import { CancelReservationResponse } from '../../models/cancelReservationResponse';
import { CheckinReservationResponse } from '../../models/checkin-reservation-response';
import { PayAnticipoResponse } from '../../models/pay-anticipo-response';
import { CreatePaymentResponse } from '../../models/create-payment-response';
import { CancelReservationModal } from '../cancel-reservation-modal/cancel-reservation-modal';
import { PayAnticipoModal } from '../pay-anticipo-modal/pay-anticipo-modal';
import { PaymentsModal } from '../payments-modal/payments-modal';

@Component({
  selector: 'app-reservation-card',
  imports: [CancelReservationModal, PayAnticipoModal, PaymentsModal, RouterLink, FormsModule],
  templateUrl: './reservation-card.html',
})
export class ReservationCard {
  private reservationService = inject(ReservationService);
  private auth = inject(AuthService);

  @Input({ required: true }) reservation!: Reservation;
  // Muestra la acción de check-in (solo tiene sentido para el anfitrión).
  @Input() showCheckin = false;
  @Output() cancelled = new EventEmitter<CancelReservationResponse>();
  @Output() checkedIn = new EventEmitter<CheckinReservationResponse>();
  @Output() paid = new EventEmitter<PayAnticipoResponse>();
  @Output() paymentRegistered = new EventEmitter<CreatePaymentResponse>();

  showModal = signal(false);
  isCancelling = signal(false);
  isCheckingIn = signal(false);
  showPayModal = signal(false);
  isPaying = signal(false);
  showPaymentsModal = signal(false);
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

  confirmPay(payload: { amount: number; method: string }): void {
    this.isPaying.set(true);
    this.reservationService.payAnticipo(this.reservation.id, payload).subscribe({
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

  openPaymentsModal(): void {
    if (!this.canViewPayments) return;
    this.showPaymentsModal.set(true);
  }

  closePaymentsModal(): void {
    this.showPaymentsModal.set(false);
  }

  onPaymentRegistered(res: CreatePaymentResponse): void {
    this.paymentRegistered.emit(res);
  }

  get isCancelled(): boolean {
    return this.reservation.status === 'cancelada';
  }

  get canCheckin(): boolean {
    return this.showCheckin && this.isConfirmed && this.isWithinStay && !this.isCheckingIn();
  }

  // El check-in solo se ofrece dentro del periodo de la reserva: ni días
  // antes del inicio ni días después del fin.
  get isWithinStay(): boolean {
    const start = this.parse(this.reservation.startDate);
    const end = this.parse(this.reservation.endDate);
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
      return false;
    }
    start.setHours(0, 0, 0, 0);
    end.setHours(0, 0, 0, 0);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return today >= start && today <= end;
  }

  // Saldo pendiente según lo reportado por el backend (0 si no viene).
  get remainingBalance(): number {
    const paid = Number(this.reservation.paidTotal ?? 0);
    return Math.max(0, Math.round((this.reservation.totalPrice - paid) * 100) / 100);
  }

  cashAmount: number | null = null;

  registerCheckin(): void {
    if (!this.canCheckin) {
      return;
    }
    const cash = Number(this.cashAmount ?? 0);
    if (this.remainingBalance > 0 && cash > 0 && cash !== this.remainingBalance) {
      this.errorMessage.set(`El efectivo debe cubrir exactamente el saldo pendiente (${this.formatPrice(this.remainingBalance)}).`);
      return;
    }
    if (!confirm(`¿Registrar el check-in de la reservación #${this.reservation.id}? Se valida que el pago esté completo y que esté dentro de las fechas.`)) {
      return;
    }
    this.isCheckingIn.set(true);
    this.errorMessage.set(null);
    this.reservationService.checkin(this.reservation.id, cash).subscribe({
      next: (res: CheckinReservationResponse) => {
        this.isCheckingIn.set(false);
        this.cashAmount = null;
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

  get isPending(): boolean {
    return this.reservation.status === 'pendiente';
  }

  // Se puede consultar el historial de pagos salvo en reservaciones canceladas.
  get canViewPayments(): boolean {
    return !this.isCancelled;
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

  // Agregar pago (abono): reservación confirmada o en curso con saldo
  // pendiente. Lo permite el backend a quien reservó, al anfitrión o admin.
  get canAddPayment(): boolean {
    return (this.isConfirmed || this.isInProgress) && this.remainingBalance > 0;
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
