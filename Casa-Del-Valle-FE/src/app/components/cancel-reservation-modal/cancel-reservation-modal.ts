import { Component, EventEmitter, HostListener, Input, Output } from '@angular/core';
import { Reservation } from '../../models/reservation';

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

  formatPrice(value: number): string {
    return `Q ${value.toLocaleString('es-GT', { minimumFractionDigits: 2 })}`;
  }
}