import { Component, Input, OnInit, Output, EventEmitter, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MaintenanceService } from '../../services/maintenance.service';
import { AffectedReservation, Maintenance } from '../../models/maintenance';

@Component({
  selector: 'app-cabin-maintenance',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './cabin-maintenance.html',
})
export class CabinMaintenance implements OnInit {
  @Input({ required: true }) cabinId = 0;
  @Output() changed = new EventEmitter<void>();
  private readonly service = inject(MaintenanceService);
  readonly records = signal<Maintenance[]>([]);
  readonly affected = signal<AffectedReservation[]>([]);
  readonly busy = signal(false);
  readonly loading = signal(false);
  readonly reviewed = signal(false);
  readonly error = signal('');
  readonly notice = signal('');
  startDate = '';
  endDate = '';
  reason = '';

  get today(): string {
    return new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Guatemala', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());
  }

  ngOnInit(): void { this.load(); }

  load(): void {
    this.loading.set(true);
    this.service.list(this.cabinId).subscribe({
      next: (records) => { this.records.set(records); this.loading.set(false); },
      error: (err) => { this.error.set(err.error?.message || 'No se pudieron cargar los mantenimientos.'); this.loading.set(false); },
    });
  }

  invalidate(): void {
    this.reviewed.set(false);
    this.affected.set([]);
    this.error.set('');
    this.notice.set('');
  }

  preview(): void {
    if (this.busy()) return;
    this.invalidate();
    if (!this.startDate || !this.endDate || this.startDate < this.today || this.endDate <= this.startDate) {
      this.error.set('Selecciona un rango válido desde hoy, con fecha de fin posterior al inicio.');
      return;
    }
    this.busy.set(true);
    this.service.preview(this.cabinId, { start_date: this.startDate, end_date: this.endDate }).subscribe({
      next: (result) => { this.affected.set(result.reservations); this.reviewed.set(true); this.busy.set(false); },
      error: (err) => { this.error.set(err.error?.message || 'No se pudieron revisar las reservas afectadas.'); this.busy.set(false); },
    });
  }

  canSubmit(): boolean {
    return this.reviewed() && !!this.reason.trim() && this.reason.trim().length <= 500 && !this.busy();
  }

  create(): void {
    if (!this.canSubmit()) return;
    this.busy.set(true);
    this.error.set('');
    this.service.create(this.cabinId, {
      start_date: this.startDate, end_date: this.endDate, reason: this.reason.trim(),
    }).subscribe({
      next: () => {
        this.busy.set(false);
        this.invalidate();
        this.startDate = this.endDate = this.reason = '';
        this.notice.set('Mantenimiento registrado. Se notificó a los huéspedes para que elijan nuevas fechas o el reembolso total.');
        this.load();
        this.changed.emit();
      },
      error: (err) => {
        this.busy.set(false);
        this.reviewed.set(false);
        this.error.set(err.error?.message || 'No se pudo registrar el mantenimiento. Revisa las reservas nuevamente.');
      },
    });
  }

  cancel(record: Maintenance): void {
    if (this.busy() || record.cancelled_at) return;
    this.busy.set(true);
    this.error.set('');
    this.notice.set('');
    this.service.cancel(this.cabinId, record.id).subscribe({
      next: () => {
        this.busy.set(false);
        this.invalidate();
        this.notice.set('Mantenimiento cancelado. Se liberó el bloqueo y se notificó a los huéspedes que recuperaron sus fechas originales.');
        this.load();
        this.changed.emit();
      },
      error: (err) => { this.busy.set(false); this.error.set(err.error?.message || 'No se pudo cancelar el mantenimiento.'); },
    });
  }

  status(record: Maintenance): string {
    if (record.cancelled_at) return 'Cancelado';
    if (record.end_date <= this.today) return 'Finalizado';
    return record.start_date > this.today ? 'Programado' : 'En mantenimiento';
  }
}
