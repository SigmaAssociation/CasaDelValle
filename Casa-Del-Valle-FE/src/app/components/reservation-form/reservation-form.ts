import { HttpErrorResponse } from '@angular/common/http';
import { Component, Input, OnInit, signal, DestroyRef, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { catchError, debounceTime, distinctUntilChanged, map, of, switchMap, tap } from 'rxjs';
import { SeasonService } from '../../services/season.service';
import { PriceQuote } from '../../models/price-quote';
import { AbstractControl, FormBuilder, FormGroup, ReactiveFormsModule, ValidationErrors, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { ReservationRequest } from '../../models/create-reservation';
import { CreateReservationResponse } from '../../models/create-reservation-response';
import { DateRange } from '../../models/date-range';
import { AuthService } from '../../services/auth.service';
import { CabinAvailabilityService } from '../../services/cabin-availability.service';
import { ReservationService } from '../../services/reservation.service';
import { AvailabilityCalendar, CalendarRange } from '../availability-calendar/availability-calendar';

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
const MS_PER_DAY = 24 * 60 * 60 * 1000;

type DateFieldName = 'start_date' | 'end_date';

@Component({
  selector: 'app-reservation-form',
  standalone: true,
  imports: [ReactiveFormsModule, RouterLink, AvailabilityCalendar],
  templateUrl: './reservation-form.html',
  styleUrl: './reservation-form.css'
})
export class ReservationForm implements OnInit {
  @Input() cabinId = 0;
  @Input() nightlyPrice = 0;

  reservationForm!: FormGroup;
  private readonly seasons = inject(SeasonService);
  private readonly destroyRef = inject(DestroyRef);
  readonly quote = signal<PriceQuote | null>(null);
  readonly quoteLoading = signal(false);
  readonly quoteError = signal('');

  isSubmitting = signal(false);
  isSuccess = signal(false);
  isError = signal(false);
  errorMessage = signal('');
  reservationId = signal<number | null>(null);
  blockedRanges = signal<DateRange[]>([]);
  isLoadingAvailability = signal(false);
  availabilityError = signal('');

  constructor(
    private formBuilder: FormBuilder,
    private reservationService: ReservationService,
    private authService: AuthService,
    private cabinAvailabilityService: CabinAvailabilityService
  ) { }

  ngOnInit(): void {
    this.reservationForm = this.formBuilder.group({
      start_date: ['', Validators.required],
      end_date: ['', Validators.required]
    }, {
      validators: (control: AbstractControl): ValidationErrors | null => this.validateDateRange(control)
    });

    // Al cambiar la fecha de inicio se revalúa la fecha de fin (min dinámico y rango).
    this.reservationForm.get('start_date')?.valueChanges.subscribe(() => {
      this.reservationForm.get('end_date')?.updateValueAndValidity();
    });

    this.loadAvailability();
    this.reservationForm.valueChanges.pipe(
      map((v) => ({ start: v.start_date || '', end: v.end_date || '' })),
      distinctUntilChanged((a, b) => a.start === b.start && a.end === b.end),
      tap(() => { this.quote.set(null); this.quoteError.set(''); this.quoteLoading.set(true); }),
      debounceTime(250),
      switchMap(({ start, end }) => {
        if (!this.parseDate(start) || !this.parseDate(end) || end <= start) { this.quoteLoading.set(false); return of({ quote: null, start, end }); }
        return this.seasons.quote(this.cabinId, start, end).pipe(
          catchError(() => { if (start === this.startDateValue && end === this.endDateValue) this.quoteError.set('No se pudo cotizar el precio por temporada. Reintenta seleccionando las fechas.'); return of(null); }),
          map((quote) => ({ quote, start, end })),
        );
      }),
      takeUntilDestroyed(this.destroyRef),
    ).subscribe(({quote, start, end}) => {
      if (start !== this.startDateValue || end !== this.endDateValue) return;
      this.quote.set(quote); this.quoteLoading.set(false);
    });
  }

  get minStartDate(): string {
    return this.toDateString(new Date());
  }

  get startDateValue(): string {
    return this.getRawValue('start_date');
  }

  get endDateValue(): string {
    return this.getRawValue('end_date');
  }

  onRangeChange(range: CalendarRange): void {
    this.reservationForm.patchValue({
      start_date: range.start ?? '',
      end_date: range.end ?? ''
    });
    this.reservationForm.get('start_date')?.markAsTouched();
    this.reservationForm.get('end_date')?.markAsTouched();
    this.reservationForm.updateValueAndValidity();
  }

  private loadAvailability(): void {
    if (this.cabinId <= 0) {
      return;
    }

    this.isLoadingAvailability.set(true);
    this.availabilityError.set('');

    this.cabinAvailabilityService.getAvailability(this.cabinId).subscribe({
      next: (availability) => {
        this.blockedRanges.set(availability.blockedRanges || []);
        this.isLoadingAvailability.set(false);
        this.reservationForm.updateValueAndValidity();
      },
      error: () => {
        this.blockedRanges.set([]);
        this.isLoadingAvailability.set(false);
        this.availabilityError.set('No se pudo cargar la disponibilidad. Verifica las fechas antes de reservar.');
      }
    });
  }

  get nights(): number | null {
    const start = this.parseDate(this.getRawValue('start_date'));
    const end = this.parseDate(this.getRawValue('end_date'));
    if (!start || !end) {
      return null;
    }
    const diff = Math.round((end.getTime() - start.getTime()) / MS_PER_DAY);
    return diff > 0 ? diff : null;
  }

  get estimatedTotal(): number | null { return this.quote()?.totalPrice ?? null; }

  get startRangeError(): string | null {
    const errors = this.reservationForm?.errors;
    if (!errors) {
      return null;
    }
    if (errors['startInPast']) {
      return 'La fecha de inicio no puede ser anterior a hoy.';
    }
    if (errors['invalidDate']) {
      return 'Ingresa una fecha válida.';
    }
    return null;
  }

  get endRangeError(): string | null {
    const errors = this.reservationForm?.errors;
    if (!errors) {
      return null;
    }
    if (errors['endBeforeStart']) {
      return 'La fecha de fin debe ser posterior a la fecha de inicio.';
    }
    if (errors['datesBlocked']) {
      return 'Algunas fechas seleccionadas ya no están disponibles. Por favor, elige otras fechas.';
    }
    if (errors['invalidDate']) {
      return 'Ingresa una fecha válida.';
    }
    return null;
  }

  submit(): void {
    if (this.reservationForm.invalid || !this.quote() || this.quoteLoading()) {
      this.reservationForm.markAllAsTouched();
      return;
    }

    const userId = this.authService.getCurrentUserId();
    if (userId === null) {
      this.errorMessage.set('Debes iniciar sesión para registrar una reservación.');
      this.isError.set(true);
      return;
    }

    if (!this.cabinId || this.cabinId <= 0) {
      this.errorMessage.set('No se pudo identificar la cabaña a reservar.');
      this.isError.set(true);
      return;
    }

    const payload: ReservationRequest = {
      user_id: userId,
      cabin_id: this.cabinId,
      start_date: this.getRawValue('start_date'),
      end_date: this.getRawValue('end_date')
    };

    this.isSubmitting.set(true);
    this.reservationService.create(payload).subscribe({
      next: (response: CreateReservationResponse) => {
        this.reservationId.set(response.reservation_id ?? null);
        this.isSubmitting.set(false);
        this.isSuccess.set(true);
        this.reservationForm.reset();
        this.loadAvailability();
      },
      error: (error: HttpErrorResponse) => {
        this.errorMessage.set(this.extractErrorMessage(error));
        this.isSubmitting.set(false);
        this.isError.set(true);
      }
    });
  }

  closeSuccess(): void {
    this.isSuccess.set(false);
    this.reservationId.set(null);
  }

  closeError(): void {
    this.isError.set(false);
  }

  formatPrice(price: number | null): string {
    return price === null ? 'Q—' : `Q${price.toFixed(2)}`;
  }

  private validateDateRange(control: AbstractControl): ValidationErrors | null {
    const group = control as FormGroup;
    const rawStart = group.get('start_date')?.value;
    const rawEnd = group.get('end_date')?.value;

    if (!rawStart || !rawEnd) {
      return null;
    }

    const start = this.parseDate(rawStart);
    const end = this.parseDate(rawEnd);
    if (!start || !end) {
      return { invalidDate: true };
    }
    if (start < this.startOfToday()) {
      return { startInPast: true };
    }
    if (end.getTime() <= start.getTime()) {
      return { endBeforeStart: true };
    }

    const blocked = this.blockedRanges();
    for (const range of blocked) {
      const blockedStart = this.parseDate(range.start);
      const blockedEnd = this.parseDate(range.end);
      if (!blockedStart || !blockedEnd) {
        continue;
      }
      if (start < blockedEnd && end > blockedStart) {
        return { datesBlocked: true };
      }
    }

    return null;
  }

  private parseDate(value: string): Date | null {
    if (!value || !DATE_PATTERN.test(value)) {
      return null;
    }
    const [year, month, day] = value.split('-').map(Number);
    const date = new Date(year, month - 1, day);
    // Verifica que la fecha no haya "rodado" (ej. 2026-02-31).
    if (
      date.getFullYear() !== year ||
      date.getMonth() !== month - 1 ||
      date.getDate() !== day
    ) {
      return null;
    }
    return date;
  }

  private startOfToday(): Date {
    const now = new Date();
    return new Date(now.getFullYear(), now.getMonth(), now.getDate());
  }

  private toDateString(date: Date): string {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  private getRawValue(field: DateFieldName): string {
    return this.reservationForm?.getRawValue()?.[field] ?? '';
  }

  private extractErrorMessage(error: HttpErrorResponse): string {
    if (error.status === 401) {
      return 'Tu sesión ha expirado. Inicia sesión nuevamente para reservar.';
    }

    if (error.error && typeof error.error === 'object' && error.error.message) {
      return error.error.message;
    }

    if (typeof error.error === 'string' && error.error.trim().length > 0) {
      return error.error;
    }

    return 'Ocurrió un error inesperado al registrar la reservación.';
  }
}
