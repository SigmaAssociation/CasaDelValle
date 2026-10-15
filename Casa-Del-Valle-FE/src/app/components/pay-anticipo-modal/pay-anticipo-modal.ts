import { Component, EventEmitter, HostListener, Input, OnInit, Output } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Reservation } from '../../models/reservation';

export interface PayAnticipoPayload {
  amount: number;
  method: string;
}

@Component({
  selector: 'app-pay-anticipo-modal',
  imports: [FormsModule],
  templateUrl: './pay-anticipo-modal.html',
})
export class PayAnticipoModal implements OnInit {
  @Input({ required: true }) reservation!: Reservation;
  @Input() loading = false;
  @Output() confirm = new EventEmitter<PayAnticipoPayload>();
  @Output() dismiss = new EventEmitter<void>();

  readonly anticipoPercent = 30;
  readonly methods = [
    { value: 'tarjeta', label: 'Tarjeta (simulada)' },
    { value: 'transferencia', label: 'Transferencia (simulada)' },
    { value: 'efectivo', label: 'Efectivo' },
  ];

  amount = 0;
  method = 'tarjeta';
  formError: string | null = null;

  @HostListener('document:keydown.escape')
  onEscape(): void {
    if (!this.loading) this.dismiss.emit();
  }

  ngOnInit(): void {
    this.amount = this.anticipoAmount;
  }

  get anticipoAmount(): number {
    return Math.round(this.reservation.totalPrice * 0.3 * 100) / 100;
  }

  get isValid(): boolean {
    return (
      Number.isFinite(this.amount) &&
      this.amount >= this.anticipoAmount &&
      this.amount <= this.reservation.totalPrice &&
      this.methods.some((m) => m.value === this.method)
    );
  }

  submit(): void {
    if (!this.isValid || this.loading) {
      this.formError = `El monto debe estar entre ${this.formatPrice(this.anticipoAmount)} y ${this.formatPrice(this.reservation.totalPrice)}.`;
      return;
    }
    this.formError = null;
    this.confirm.emit({ amount: this.amount, method: this.method });
  }

  formatDate(value: string): string {
    const d = new Date(value.length === 10 ? `${value}T00:00:00` : value);
    return d.toLocaleDateString('es-GT', { day: 'numeric', month: 'long', year: 'numeric' });
  }

  formatPrice(value: number): string {
    return `Q ${Number(value).toLocaleString('es-GT', { minimumFractionDigits: 2 })}`;
  }
}
