import { Component, EventEmitter, Input, OnInit, Output, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MaintenanceService } from '../../services/maintenance.service';
import { PendingMaintenance } from '../../models/maintenance';
import { AvailabilityCalendar, CalendarRange } from '../availability-calendar/availability-calendar';

@Component({
  selector: 'app-maintenance-decision', standalone: true,
  imports: [CommonModule, AvailabilityCalendar], templateUrl: './maintenance-decision.html',
})
export class MaintenanceDecision implements OnInit {
  @Input({ required: true }) reservationId = 0;
  @Output() resolved = new EventEmitter<string>();
  private readonly service = inject(MaintenanceService);
  readonly details = signal<PendingMaintenance | null>(null);
  readonly loading = signal(false);
  readonly busy = signal(false);
  readonly error = signal('');
  readonly option = signal<'dates' | 'refund' | null>(null);
  readonly start = signal('');
  readonly end = signal('');
  get today(): string {
    return new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Guatemala', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());
  }
  ngOnInit(): void { this.load(); }
  load(): void {
    this.loading.set(true); this.error.set('');
    this.service.pending(this.reservationId).subscribe({
      next: (d) => { this.details.set(d); this.loading.set(false); },
      error: (err) => { this.error.set(err.error?.message || 'No se pudieron cargar tus opciones.'); this.loading.set(false); },
    });
  }
  rangeChanged(range: CalendarRange): void { this.start.set(range.start || ''); this.end.set(range.end || ''); }
  get selectedNights(): number { return Math.round((Date.parse(this.end()) - Date.parse(this.start())) / 86400000); }
  canReschedule(): boolean {
    const d = this.details();
    return !!d && !this.busy() && this.start() >= this.today && this.selectedNights === d.nights &&
      !d.blocked_ranges.some((b) => this.start() < b.end && this.end() > b.start);
  }
  decide(): void {
    if (this.busy() || !this.details() || !this.option()) return;
    if (this.option() === 'dates' && !this.canReschedule()) return;
    this.busy.set(true); this.error.set('');
    const decision = this.option() === 'refund' ? { action: 'reembolsar' as const } :
      { action: 'reprogramar' as const, start_date: this.start(), end_date: this.end() };
    this.service.resolve(this.reservationId, decision).subscribe({
      next: (result) => { this.busy.set(false); this.resolved.emit(result.message); },
      error: (err) => {
        this.busy.set(false); this.error.set(err.error?.message || 'No se pudo guardar tu decisión.');
        this.start.set(''); this.end.set('');
      },
    });
  }
}
