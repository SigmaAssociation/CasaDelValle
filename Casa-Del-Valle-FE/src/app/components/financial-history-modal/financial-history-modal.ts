import { Component, EventEmitter, HostListener, Input, Output } from '@angular/core';
import { Reservation } from '../../models/reservation';
import { FinancialMovement, RefundsHistoryResponse } from '../../models/refund';

@Component({
  selector: 'app-financial-history-modal',
  templateUrl: './financial-history-modal.html',
})
export class FinancialHistoryModal {
  @Input({ required: true }) reservation!: Reservation;
  @Input() history: RefundsHistoryResponse | null = null;
  @Input() loading = false;
  @Input() errorMessage: string | null = null;

  @Output() dismiss = new EventEmitter<void>();
  @Output() openRefund = new EventEmitter<void>();

  @HostListener('document:keydown.escape')
  onEscape(): void {
    this.dismiss.emit();
  }

  movementLabel(movement: FinancialMovement): string {
    switch (movement.subtype) {
      case 'anticipo':
        return 'Anticipo';
      case 'reembolso':
        return 'Reembolso';
      default:
        return movement.type === 'reembolso' ? 'Reembolso' : 'Pago';
    }
  }

  formatPrice(value: number): string {
    return `Q ${value.toLocaleString('es-GT', { minimumFractionDigits: 2 })}`;
  }

  formatDateTime(value: string): string {
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) {
      return value;
    }
    return d.toLocaleString('es-GT', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }
}
