import { HttpErrorResponse } from '@angular/common/http';
import { Component, Input, OnInit, signal } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, ReactiveFormsModule, ValidationErrors, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { ReservationRequest } from '../../models/create-reservation';
import { CreateReservationResponse } from '../../models/create-reservation-response';
import { AuthService } from '../../services/auth.service';
import { ReservationService } from '../../services/reservation.service';

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
const MS_PER_DAY = 24 * 60 * 60 * 1000;

type DateFieldName = 'start_date' | 'end_date';

@Component({
  selector: 'app-reservation-form',
  standalone: true,
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './reservation-form.html',
  styleUrl: './reservation-form.css'
})
export class ReservationForm implements OnInit {
  @Input() cabinId = 0;
  @Input() nightlyPrice = 0;

  reservationForm!: FormGroup;

  isSubmitting = signal(false);
  isSuccess = signal(false);
  isError = signal(false);
  errorMessage = signal('');
  reservationId = signal<number | null>(null);

  constructor(
    private formBuilder: FormBuilder,
    private reservationService: ReservationService,
    private authService: AuthService
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
  }

  get minStartDate(): string {
    return this.toDateString(new Date());
  }

  get minEndDate(): string | null {
    const start = this.getRawValue('start_date');
    const parsedStart = start ? this.parseDate(start) : null;
    if (!parsedStart) {
      return null;
    }
    return this.toDateString(new Date(parsedStart.getTime() + MS_PER_DAY));
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

  get estimatedTotal(): number | null {
    const nights = this.nights;
    if (nights === null || !Number.isFinite(this.nightlyPrice)) {
      return null;
    }
    return nights * this.nightlyPrice;
  }

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
    if (errors['invalidDate']) {
      return 'Ingresa una fecha válida.';
    }
    return null;
  }

  fieldError(field: DateFieldName): string | null {
    const control = this.reservationForm.get(field);
    if (!control || !control.invalid || !(control.dirty || control.touched)) {
      return null;
    }
    if (control.errors?.['required']) {
      return field === 'start_date'
        ? 'La fecha de inicio es obligatoria.'
        : 'La fecha de fin es obligatoria.';
    }
    if (control.errors?.['invalidDate']) {
      return 'Ingresa una fecha válida.';
    }
    return null;
  }

  isInvalid(field: DateFieldName): boolean {
    const control = this.reservationForm.get(field);
    return !!(control && control.invalid && (control.dirty || control.touched));
  }

  submit(): void {
    if (this.reservationForm.invalid) {
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
