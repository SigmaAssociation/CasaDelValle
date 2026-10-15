import { Component, EventEmitter, HostListener, Input, OnInit, Output, inject, signal } from '@angular/core';
import { Reservation } from '../../models/reservation';
import { PaymentService } from '../../services/payment.service';
import { ReservationPayments } from '../../models/reservation-payments';
import { CreatePaymentResponse } from '../../models/create-payment-response';
import { PAYMENT_METHODS, PaymentMethod, paymentMethodLabel } from '../../models/payment-method';

@Component({
  selector: 'app-payments-modal',
  templateUrl: './payments-modal.html',
})
export class PaymentsModal implements OnInit {
  private paymentService = inject(PaymentService);

  @Input({ required: true }) reservation!: Reservation;
  @Output() registered = new EventEmitter<CreatePaymentResponse>();
  @Output() dismiss = new EventEmitter<void>();

  readonly methods = PAYMENT_METHODS;

  summary = signal<ReservationPayments | null>(null);
  isLoading = signal(true);
  loadError = signal(false);

  amount = signal<number | null>(null);
  method = signal<PaymentMethod>('tarjeta');
  isSubmitting = signal(false);
  formError = signal<string | null>(null);

  ngOnInit(): void {
    this.load();
  }

  @HostListener('document:keydown.escape')
  onEscape(): void {
    this.close();
  }

  load(): void {
    this.isLoading.set(true);
    this.loadError.set(false);
    this.paymentService.getByReservation(this.reservation.id).subscribe({
      next: (res) => {
        this.summary.set(res);
        this.isLoading.set(false);
      },
      error: () => {
        this.isLoading.set(false);
        this.loadError.set(true);
      },
    });
  }

  // Solo se puede registrar en reservaciones confirmadas o en curso.
  get canRegister(): boolean {
    return this.reservation.status === 'confirmada' || this.reservation.status === 'en_curso';
  }

  onAmountInput(event: Event): void {
    const value = (event.target as HTMLInputElement).value;
    this.amount.set(value === '' ? null : Number(value));
  }

  onMethodChange(event: Event): void {
    this.method.set((event.target as HTMLSelectElement).value as PaymentMethod);
  }

  submit(): void {
    if (!this.canRegister || this.isSubmitting()) return;

    const amount = this.amount();
    if (amount == null || Number.isNaN(amount) || amount <= 0) {
      this.formError.set('El monto debe ser mayor que 0.');
      return;
    }

    const balance = this.summary()?.balance;
    if (balance != null && amount > balance) {
      this.formError.set('El monto no puede superar el saldo pendiente.');
      return;
    }

    this.isSubmitting.set(true);
    this.formError.set(null);

    this.paymentService.create(this.reservation.id, { amount, method: this.method() }).subscribe({
      next: (res) => {
        this.isSubmitting.set(false);
        this.summary.set({
          reservationId: res.reservationId,
          total: res.total,
          paid: res.paid,
          balance: res.balance,
          payments: res.payments,
        });
        this.amount.set(null);
        this.registered.emit(res);
      },
      error: (err) => {
        this.isSubmitting.set(false);
        this.formError.set(
          err.error?.message ?? err.error?.error ?? 'No se pudo registrar el pago. Intenta de nuevo.'
        );
      },
    });
  }

  close(): void {
    if (this.isSubmitting()) return;
    this.dismiss.emit();
  }

  methodLabel(method: string): string {
    return paymentMethodLabel(method);
  }

  formatDate(value: string): string {
    const d = new Date(value.length === 10 ? `${value}T00:00:00` : value);
    return d.toLocaleDateString('es-GT', { day: 'numeric', month: 'long', year: 'numeric' });
  }

  formatPrice(value: number): string {
    return `Q ${value.toLocaleString('es-GT', { minimumFractionDigits: 2 })}`;
  }
}
