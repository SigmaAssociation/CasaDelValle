import { DatePipe } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, signal } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { Season, SeasonRequest } from '../../models/season';
import { SeasonService } from '../../services/season.service';

@Component({
  selector: 'app-admin-seasons-page',
  standalone: true,
  imports: [ReactiveFormsModule, DatePipe],
  templateUrl: './admin-seasons-page.html',
})
export class AdminSeasonsPage implements OnInit {
  seasons = signal<Season[]>([]);
  loading = signal(false);
  error = signal('');
  success = signal('');
  saving = signal(false);
  formError = signal('');
  modalOpen = signal(false);
  editingId = signal<number | null>(null);

  form!: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private seasonService: SeasonService,
  ) {}

  ngOnInit(): void {
    this.form = this.formBuilder.group(
      {
        name: ['', [Validators.required, Validators.maxLength(100)]],
        suggestedAdjustmentPercent: [
          0,
          [Validators.required, Validators.min(0), Validators.max(100)],
        ],
        startDate: ['', [Validators.required]],
        endDate: ['', [Validators.required]],
        active: [true],
      },
      { validators: this.dateRangeValidator },
    );
    this.loadSeasons();
  }

  private dateRangeValidator(group: AbstractControl): ValidationErrors | null {
    const start = group.get('startDate')?.value;
    const end = group.get('endDate')?.value;
    return start && end && start > end ? { dateRange: true } : null;
  }

  loadSeasons(): void {
    this.loading.set(true);
    this.error.set('');
    this.seasonService.list(true).subscribe({
      next: (seasons) => {
        this.seasons.set(seasons);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.seasons.set([]);
        this.error.set(err.error?.message ?? 'No se pudieron cargar las temporadas.');
        this.loading.set(false);
      },
    });
  }

  openCreate(): void {
    this.editingId.set(null);
    this.formError.set('');
    this.success.set('');
    this.form.reset({
      name: '',
      suggestedAdjustmentPercent: 0,
      startDate: '',
      endDate: '',
      active: true,
    });
    this.modalOpen.set(true);
  }

  openEdit(season: Season): void {
    this.editingId.set(season.id);
    this.formError.set('');
    this.success.set('');
    this.form.reset({
      name: season.name,
      suggestedAdjustmentPercent: season.suggestedAdjustmentPercent,
      startDate: season.startDate,
      endDate: season.endDate,
      active: season.active,
    });
    this.modalOpen.set(true);
  }

  closeModal(): void {
    this.modalOpen.set(false);
    this.formError.set('');
  }

  isInvalid(controlName: string): boolean {
    const control = this.form.get(controlName);
    return !!(control && control.invalid && (control.dirty || control.touched));
  }

  save(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const value = this.form.getRawValue();
    const request: SeasonRequest = {
      name: value.name.trim(),
      suggestedAdjustmentPercent: Number(value.suggestedAdjustmentPercent),
      active: !!value.active,
      startDate: value.startDate,
      endDate: value.endDate,
    };

    this.saving.set(true);
    this.formError.set('');
    const id = this.editingId();
    const onSuccess = () => {
      this.saving.set(false);
      this.modalOpen.set(false);
      this.success.set(id === null ? 'Temporada creada.' : 'Temporada actualizada.');
      this.loadSeasons();
    };
    const onError = (err: HttpErrorResponse) => {
      this.saving.set(false);
      this.formError.set(err.error?.message ?? 'No se pudo guardar la temporada.');
    };

    if (id === null) {
      this.seasonService.create(request).subscribe({ next: onSuccess, error: onError });
    } else {
      this.seasonService.update(id, request).subscribe({ next: onSuccess, error: onError });
    }
  }

  deactivate(season: Season): void {
    if (!confirm(`¿Desactivar la temporada "${season.name}"? Podrás reactivarla editándola.`)) {
      return;
    }
    this.error.set('');
    this.success.set('');
    this.seasonService.deactivate(season.id).subscribe({
      next: () => {
        this.success.set('Temporada desactivada.');
        this.loadSeasons();
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.message ?? 'No se pudo desactivar la temporada.');
      },
    });
  }
}
