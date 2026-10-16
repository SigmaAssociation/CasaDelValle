import { Component, EventEmitter, inject, Input, Output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { Reservation } from '../../models/reservation';
import { ReservationService } from '../../services/reservation.service';
import { RefundService } from '../../services/refund.service';
import { AuthService } from '../../services/auth.service';
import { CancelReservationResponse } from '../../models/cancelReservationResponse';
import { CheckinReservationResponse } from '../../models/checkin-reservation-response';
import { PayAnticipoResponse } from '../../models/pay-anticipo-response';
import { CreatePaymentResponse } from '../../models/create-payment-response';
import { RefundEvent, RefundsHistoryResponse, RefundTier, anticipoFor, estimatedRefundAmount, refundTierFor } from '../../models/refund';
import { CancelReservationModal } from '../cancel-reservation-modal/cancel-reservation-modal';
import { PayAnticipoModal } from '../pay-anticipo-modal/pay-anticipo-modal';
import { PaymentsModal } from '../payments-modal/payments-modal';
import { RefundModal } from '../refund-modal/refund-modal';
import { FinancialHistoryModal } from '../financial-history-modal/financial-history-modal';

@Component({
  selector: 'app-reservation-card',
  imports: [CancelReservationModal, PayAnticipoModal, PaymentsModal, RefundModal, FinancialHistoryModal, RouterLink, FormsModule],
  templateUrl: './reservation-card.html',
})
export class ReservationCard {
  private reservationService = inject(ReservationService);
  private refundService = inject(RefundService);
  private auth = inject(AuthService);

  @Input({ required: true }) reservation!: Reservation;
  // Muestra la acción de check-in (solo tiene sentido para el anfitrión).
  @Input() showCheckin = false;
  @Output() cancelled = new EventEmitter<CancelReservationResponse>();
  @Output() checkedIn = new EventEmitter<CheckinReservationResponse>();
  @Output() paid = new EventEmitter<PayAnticipoResponse>();
  @Output() paymentRegistered = new EventEmitter<CreatePaymentResponse>();
  @Output() refunded = new EventEmitter<RefundEvent>();

  showModal = signal(false);
  isCancelling = signal(false);
  isCheckingIn = signal(false);
  showPayModal = signal(false);
  isPaying = signal(false);
  showPaymentsModal = signal(false);
  errorMessage = signal<string | null>(null);
  statusMessage = signal<string | null>(null);

  showRefundModal = signal(false);
  showHistoryModal = signal(false);
  isLoadingHistory = signal(false);
  isRefunding = signal(false);
  historyError = signal<string | null>(null);
  refundHistory = signal<RefundsHistoryResponse | null>(null);

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
        this.errorMessage.set(null);
        this.statusMessage.set(
          res.refund
            ? `Se registró un reembolso ${res.refund.kind} de ${this.formatPrice(res.refund.amount)}.`
            : this.refundTier === 'ninguno'
              ? 'Reservación cancelada. Por cancelar con menos de 3 días de anticipación no corresponde reembolso.'
              : res.message
        );
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

  openHistoryModal(): void {
    if (!this.canViewHistory) return;
    this.showHistoryModal.set(true);
    this.loadHistory();
  }

  closeHistoryModal(): void {
    this.showHistoryModal.set(false);
  }

  openRefundModal(): void {
    if (!this.canRefund) return;
    this.errorMessage.set(null);
    this.showHistoryModal.set(false);
    this.showRefundModal.set(true);
    this.loadHistory();
  }

  closeRefundModal(): void {
    if (this.isRefunding()) return;
    this.showRefundModal.set(false);
  }

  confirmRefund(amount?: number): void {
    this.isRefunding.set(true);
    this.errorMessage.set(null);
    this.refundService
      .create(this.reservation.id, amount === undefined ? {} : { amount })
      .subscribe({
        next: (res) => {
          this.isRefunding.set(false);
          this.showRefundModal.set(false);
          this.statusMessage.set(res.message);
          this.loadHistory(() => {
            const history = this.refundHistory();
            if (history) {
              this.refunded.emit({
                message: res.message,
                reservationId: this.reservation.id,
                history,
              });
            }
          });
        },
        error: (err) => {
          this.isRefunding.set(false);
          this.errorMessage.set(
            err.error?.message ?? err.error?.error ?? 'No se pudo registrar el reembolso. Intenta de nuevo.'
          );
        },
      });
  }

  private loadHistory(onDone?: () => void): void {
    this.isLoadingHistory.set(true);
    this.historyError.set(null);
    this.refundService.getHistory(this.reservation.id).subscribe({
      next: (history) => {
        this.refundHistory.set(history);
        this.isLoadingHistory.set(false);
        onDone?.();
      },
      error: (err) => {
        console.error('Error al cargar el historial financiero:', err);
        this.isLoadingHistory.set(false);
        this.historyError.set('No se pudo cargar el historial financiero.');
        onDone?.();
      },
    });
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

  get isOwnReservation(): boolean {
    const userId = this.auth.getCurrentUserId();
    return userId != null && userId === this.reservation.guestId;
  }

  get isHostOfCabin(): boolean {
    const userId = this.auth.getCurrentUserId();
    return userId != null && this.reservation.hostId != null && userId === this.reservation.hostId;
  }

  get isAdminUser(): boolean {
    return this.auth.user()?.role === 1;
  }

  get canPay(): boolean {
    return this.isPending && this.isOwnReservation;
  }

  // Agregar pago (abono): reservación confirmada o en curso con saldo
  // pendiente. Lo permite el backend a quien reservó, al anfitrión o admin.
  get canAddPayment(): boolean {
    return (this.isConfirmed || this.isInProgress) && this.remainingBalance > 0;
  }

  get canRefund(): boolean {
    return this.isCancelled && (this.isHostOfCabin || this.isAdminUser);
  }

  get canViewHistory(): boolean {
    return this.isCancelled;
  }

  get anticipoAmount(): number {
    return anticipoFor(this.reservation.totalPrice);
  }

  get nights(): number {
    const ms =
      this.parse(this.reservation.endDate).getTime() -
      this.parse(this.reservation.startDate).getTime();
    return Math.round(ms / 86_400_000);
  }

  get refundTier(): RefundTier {
    return refundTierFor(this.reservation.startDate);
  }

  get refundAmount(): number {
    return estimatedRefundAmount(this.reservation);
  }

  get refundTierLabel(): string {
    switch (this.refundTier) {
      case 'total':
        return `Reembolso total: ${this.formatPrice(this.refundAmount)}`;
      case 'parcial':
        return `Reembolso parcial (70%): ${this.formatPrice(this.refundAmount)}`;
      default:
        return 'Cancelación sin reembolso (menos de 3 días)';
    }
  }

  get canCancel(): boolean {
    return this.isPending || this.isConfirmed;
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

  private parse(value: string): Date {
    return new Date(value.length === 10 ? `${value}T00:00:00` : value);
  }
}
