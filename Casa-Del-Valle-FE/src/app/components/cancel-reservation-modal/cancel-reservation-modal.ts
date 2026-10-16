import { Component, EventEmitter, HostListener, Input, Output } from '@angular/core';
import { Reservation } from '../../models/reservation';
import { RefundTier, estimatedRefundAmount, refundTierFor } from '../../models/refund';

@Component({
  selector: 'app-cancel-reservation-modal',
  templateUrl: './cancel-reservation-modal.html',
})
export class CancelReservationModal {
  @Input({ required: true }) reservation!: Reservation;
  @Input() loading = false;
  @Output() confirm = new EventEmitter<void>();
  @Output() dismiss = new EventEmitter<void>();

  @HostListener('document:keydown.escape')
  onEscape(): void {
    if (!this.loading) this.dismiss.emit();
  }

  formatDate(value: string): string {
    const d = new Date(value.length === 10 ? `${value}T00:00:00` : value);
    return d.toLocaleDateString('es-GT', { day: 'numeric', month: 'long', year: 'numeric' });
  }

  get refundTier(): RefundTier {
    return refundTierFor(this.reservation.startDate);
  }

  get refundAmount(): number {
    return estimatedRefundAmount(this.reservation);
  }

  get hasRefund(): boolean {
    return this.refundAmount > 0;
  }

  get refundMessage(): string {
    if (this.refundTier === 'ninguno') {
      return 'Al cancelar con menos de 3 días de anticipación no aplica reembolso.';
    }
    if (this.refundAmount <= 0) {
      return 'La reservación aún no tiene pagos registrados, por lo que no aplica reembolso.';
    }
    const kind = this.refundTier === 'total' ? 'total' : 'parcial (70%)';
    return `Se registrará un reembolso ${kind} de ${this.formatPrice(this.refundAmount)} al confirmar.`;
  }

  formatPrice(value: number): string {
    return `Q ${value.toLocaleString('es-GT', { minimumFractionDigits: 2 })}`;
  }
}