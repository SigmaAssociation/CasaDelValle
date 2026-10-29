import { CurrencyPipe, DatePipe } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, inject, signal } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  ValidationErrors,
  ValidatorFn,
  Validators,
} from '@angular/forms';
import { Cabin } from '../../models/cabin';
import { IncomeReportFilter, IncomeReportResponse } from '../../models/report-income';
import { AuthService } from '../../services/auth.service';
import { CabinService } from '../../services/cabin.service';
import { ReportIncomeService } from '../../services/reports-income.service';

@Component({
  selector: 'app-income-report-page',
  standalone: true,
  imports: [ReactiveFormsModule, CurrencyPipe, DatePipe],
  templateUrl: './income-report-page.html',
})
export class IncomeReportPage implements OnInit {
  private readonly formBuilder = inject(FormBuilder);
  private readonly reportService = inject(ReportIncomeService);
  private readonly cabinService = inject(CabinService);
  private readonly authService = inject(AuthService);

  filterForm!: FormGroup;
  report = signal<IncomeReportResponse | null>(null);
  cabins = signal<Cabin[]>([]);
  loading = signal(false);
  error = signal('');

  ngOnInit(): void {
    this.filterForm = this.formBuilder.group(
      {
        start_date: [this.monthStart(), [Validators.required]],
        end_date: [this.today(), [Validators.required]],
        cabin_id: [null as number | null],
      },
      { validators: this.dateRangeValidator('start_date', 'end_date') },
    );

    this.loadCabins();
    this.search();
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
    this.filterForm.reset({
      start_date: this.monthStart(),
      end_date: this.today(),
      cabin_id: null,
    });
    this.search();
  }

  typeLabel(type: string): string {
    const labels: Record<string, string> = {
      anticipo: 'Anticipo',
      pago: 'Pago',
      reembolso: 'Reembolso',
    };
    return labels[type] ?? type;
  }

  typeClass(type: string): string {
    const classes: Record<string, string> = {
      anticipo: 'bg-blue-100 text-blue-800',
      pago: 'bg-green-100 text-green-800',
      reembolso: 'bg-red-100 text-red-800',
    };
    return classes[type] ?? 'bg-gray-100 text-gray-600';
  }

  private loadCabins(): void {
    if (this.authService.isAdmin()) {
      this.cabinService.getCabins().subscribe({
        next: (cabins) => this.cabins.set(cabins),
        error: () => this.cabins.set([]),
      });
      return;
    }

    const userId = this.authService.getCurrentUserId();
    if (!userId) {
      this.cabins.set([]);
      return;
    }

    this.cabinService.getCabinsByUserId(userId).subscribe({
      next: (cabins) => this.cabins.set(cabins),
      error: () => this.cabins.set([]),
    });
  }

  private dateRangeValidator(fromKey: string, toKey: string): ValidatorFn {
    return (group: AbstractControl): ValidationErrors | null => {
      const from = group.get(fromKey)?.value;
      const to = group.get(toKey)?.value;
      return from && to && from > to ? { invalidRange: true } : null;
    };
  }

  private loadReport(): void {
    this.loading.set(true);
    this.error.set('');

    const value = this.filterForm.value;
    const filter: IncomeReportFilter = {
      startDate: value.start_date,
      endDate: value.end_date,
    };

    if (value.cabin_id != null) {
      filter.cabinId = value.cabin_id;
    }

    this.reportService.getIncomeReport(filter).subscribe({
      next: (report) => {
        this.report.set(report);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.report.set(null);
        this.error.set(err.error?.message ?? 'No se pudo cargar el reporte de ingresos.');
        this.loading.set(false);
      },
    });
  }

  private monthStart(): string {
    const now = new Date();
    return this.formatDate(new Date(now.getFullYear(), now.getMonth(), 1));
  }

  private today(): string {
    return this.formatDate(new Date());
  }

  private formatDate(date: Date): string {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }
}
