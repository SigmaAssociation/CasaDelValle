import { Component, EventEmitter, HostListener, Input, Output, computed, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Reservation } from '../../models/reservation';
import { RefundsHistoryResponse } from '../../models/refund';

@Component({
  selector: 'app-refund-modal',
  imports: [FormsModule],
  templateUrl: './refund-modal.html',
})
export class RefundModal {
  @Input({ required: true }) reservation!: Reservation;
  @Input() history: RefundsHistoryResponse | null = null;
  @Input() loading = false;
  @Input() errorMessage: string | null = null;

  @Output() confirm = new EventEmitter<number | undefined>();
  @Output() dismiss = new EventEmitter<void>();

  mode = signal<'total' | 'parcial'>('total');
  partialAmount = signal<number | null>(null);

  @HostListener('document:keydown.escape')
  onEscape(): void {
    if (!this.loading) this.dismiss.emit();
  }

  get available(): number {
    return this.round2(this.history?.refundableAvailable ?? 0);
  }

  amountToRefund = computed(() => {
    if (this.mode() === 'total') {
      return this.available;
    }
    const value = Number(this.partialAmount());
    return Number.isFinite(value) ? this.round2(value) : 0;
  });

  get validationError(): string | null {
    if (this.mode() !== 'parcial') {
      return null;
    }
    const value = this.amountToRefund();
    if (value <= 0) {
      return 'Ingresa un monto mayor a cero.';
    }
    if (value > this.available) {
      return 'El monto no puede superar el disponible a reembolsar.';
    }
    return null;
  }

  get canConfirm(): boolean {
    return !this.loading && this.available > 0 && this.validationError === null;
  }

  submit(): void {
    if (!this.canConfirm) {
      return;
    }
    this.confirm.emit(this.mode() === 'total' ? undefined : this.amountToRefund());
  }

  formatPrice(value: number): string {
    return `Q ${value.toLocaleString('es-GT', { minimumFractionDigits: 2 })}`;
  }

  private round2(value: number): number {
    return Math.round(value * 100) / 100;
  }
}
