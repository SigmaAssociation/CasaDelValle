import { CurrencyPipe, DatePipe } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, signal } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, ReactiveFormsModule, ValidationErrors, ValidatorFn, Validators } from '@angular/forms';
import { ReportReservationsService } from '../../services/reports-reservations.service';
import { ReservationStatus, ReservationsReportFilter } from '../../models/report-reservations';
import { ReservationsReportResponse } from '../../models/report-reservations';
import { CabinService } from '../../services/cabin.service';
import { Cabin } from '../../models/cabin';

@Component({
  selector: 'app-reservations-report-page',
  standalone: true,
  imports: [ReactiveFormsModule, CurrencyPipe, DatePipe],
  templateUrl: './reservations-report-page.html',
})
export class ReservationsReportPage implements OnInit {
  filterForm!: FormGroup;

  report = signal<ReservationsReportResponse | null>(null);
  cabins = signal<Cabin[]>([]);
  loading = signal(false);
  error = signal('');

  statuses: ReservationStatus[] = ['pendiente', 'confirmada', 'en_curso', 'finalizada', 'cancelada', 'por_mantenimiento'];

  constructor(
    private formBuilder: FormBuilder,
    private reportService: ReportReservationsService,
    private cabinService: CabinService
  ) { }

  ngOnInit(): void {
    this.filterForm = this.formBuilder.group(
      {
        host_id: [null as number | null, [Validators.min(1)]],
        cabin_id: [null as number | null],
        guest_id: [null as number | null, [Validators.min(1)]],
        status: [null as ReservationStatus | null],
        created_from: [null as string | null],
        created_to: [null as string | null],
      },
      {
        validators: this.dateRangeValidator('created_from', 'created_to', 'createdRange'),
      }
    );

    this.loadCabins();
    this.loadReport();
  }

  private loadCabins(): void {
    this.cabinService.getCabins().subscribe({
      next: (cabins) => this.cabins.set(cabins),
      error: () => this.cabins.set([]),
    });
  }

  private dateRangeValidator(fromKey: string, toKey: string, errorKey: string): ValidatorFn {
    return (group: AbstractControl): ValidationErrors | null => {
      const from = group.get(fromKey)?.value;
      const to = group.get(toKey)?.value;
      return from && to && from > to ? { [errorKey]: true } : null;
    };
  }


  isInvalid(controlName: string): boolean {
    const control = this.filterForm.get(controlName);
    return !!(control && control.invalid && (control.dirty || control.touched));
  }

  search(): void {
    if (this.filterForm.invalid) {
      this.filterForm.markAllAsTouched();
      return;
    }
    this.loadReport();
  }

  clear(): void {
    this.filterForm.reset();
    this.loadReport();
  }

  statusLabel(status: string): string {
    return status === 'por_mantenimiento' ? 'Afectada por mantenimiento' : status.replace('_', ' ');
  }

  statusClass(status: string): string {
    const classes: Record<string, string> = {
      pendiente: 'bg-yellow-100 text-yellow-800',
      confirmada: 'bg-green-100 text-green-800',
      en_curso: 'bg-blue-100 text-blue-800',
      finalizada: 'bg-gray-100 text-gray-600',
      cancelada: 'bg-red-100 text-red-800',
      por_mantenimiento: 'bg-orange/10 text-orange',
    };
    return classes[status] ?? 'bg-gray-100 text-gray-600';
  }

  private loadReport(): void {
    this.loading.set(true);
    this.error.set('');

    const f = this.filterForm.value;
    const filter: ReservationsReportFilter = {};
    if (f.host_id != null) filter.host_id = f.host_id;
    if (f.cabin_id != null) filter.cabin_id = f.cabin_id;
    if (f.guest_id != null) filter.guest_id = f.guest_id;
    if (f.status) filter.status = f.status;
    if (f.created_from) filter.created_from = f.created_from;
    if (f.created_to) filter.created_to = f.created_to;

    this.reportService.getReservationsReport(filter).subscribe({
      next: (report) => {
        this.report.set(report);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.report.set(null);
        this.error.set(err.error?.message ?? 'No se pudo cargar el reporte.');
        this.loading.set(false);
      },
    });
  }
}