import { Component, Input, OnChanges, SimpleChanges, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AbstractControl, FormBuilder, ReactiveFormsModule, ValidationErrors, Validators } from '@angular/forms';
import { forkJoin } from 'rxjs';
import { CabinSeasonPrice, Season } from '../../models/season';
import { SeasonService } from '../../services/season.service';

function validateSeasonPrice(control: AbstractControl): ValidationErrors | null {
  const { startDate, endDate, pricingMode, adjustmentPercent, fixedPrice } = control.value;
  const errors: ValidationErrors = {};
  if (startDate && endDate && endDate < startDate) errors['dateOrder'] = true;
  if (pricingMode === 'percent') {
    if (adjustmentPercent == null || !Number.isFinite(Number(adjustmentPercent)) || Number(adjustmentPercent) <= -100 || Number(adjustmentPercent) > 9999.99) errors['invalidPercent'] = true;
  } else if (pricingMode === 'fixed') {
    if (fixedPrice == null || !Number.isFinite(Number(fixedPrice)) || Number(fixedPrice) <= 0 || Number(fixedPrice) > 99999999.99) errors['invalidFixed'] = true;
  } else errors['invalidMode'] = true;
  return Object.keys(errors).length ? errors : null;
}

@Component({
  selector: 'app-cabin-season-prices',
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './cabin-season-prices.html',
  styleUrl: './cabin-season-prices.css',
})
export class CabinSeasonPrices implements OnChanges {
  @Input({ required: true }) cabinId!: number;
  @Input({ required: true }) basePrice = 0;
  readonly seasons = signal<Season[]>([]);
  readonly prices = signal<CabinSeasonPrice[]>([]);
  readonly loading = signal(false);
  readonly loadFailed = signal(false);
  readonly error = signal<string | null>(null);
  readonly success = signal<string | null>(null);
  readonly submitting = signal(false);
  readonly attempted = signal(false);
  readonly deleting = signal<number | null>(null);
  readonly pendingDelete = signal<number | null>(null);
  readonly form;

  constructor(fb: FormBuilder, private readonly service: SeasonService) {
    this.form = fb.group({
      seasonId: [''],
      startDate: ['', Validators.required],
      endDate: ['', Validators.required],
      pricingMode: ['percent'],
      adjustmentPercent: [0 as number | null],
      fixedPrice: [null as number | null],
    }, { validators: validateSeasonPrice });
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['cabinId'] && this.cabinId) this.load();
  }

  load(): void {
    this.loading.set(true); this.loadFailed.set(false); this.error.set(null);
    forkJoin({ seasons: this.service.list(), prices: this.service.listCabinPrices(this.cabinId) }).subscribe({
      next: ({ seasons, prices }) => { this.seasons.set(seasons); this.prices.set(prices); this.loading.set(false); },
      error: () => { this.error.set('No pudimos cargar las temporadas y las tarifas. Inténtalo de nuevo.'); this.loading.set(false); this.loadFailed.set(true); },
    });
  }

  get selectedSeason(): Season | undefined {
    return this.seasons().find((season) => season.id === Number(this.form.controls.seasonId.value));
  }

  get recommendedPrice(): number {
    return this.round(this.basePrice * (1 + (this.selectedSeason?.suggestedAdjustmentPercent ?? 0) / 100));
  }

  get previewPrice(): number | null {
    if (this.form.hasError('invalidFixed') || this.form.hasError('invalidPercent') || this.form.hasError('invalidMode')) return null;
    return this.form.controls.pricingMode.value === 'fixed'
      ? this.round(Number(this.form.controls.fixedPrice.value))
      : this.round(this.basePrice * (1 + Number(this.form.controls.adjustmentPercent.value) / 100));
  }

  selectSeason(): void {
    const season = this.selectedSeason;
    this.success.set(null);
    if (season) {
      this.form.patchValue({ startDate: season.startDate, endDate: season.endDate });
      this.useRecommendation();
    } else {
      this.form.patchValue({ startDate: '', endDate: '', adjustmentPercent: 0, fixedPrice: null, pricingMode: 'percent' });
    }
  }

  useRecommendation(): void {
    const season = this.selectedSeason;
    if (season) this.form.patchValue({ pricingMode: 'percent', adjustmentPercent: season.suggestedAdjustmentPercent, fixedPrice: null });
  }

  save(): void {
    if (this.submitting() || this.loading() || this.loadFailed()) return;
    this.attempted.set(true);
    if (this.form.invalid) { this.form.markAllAsTouched(); return; }
    const value = this.form.getRawValue();
    this.submitting.set(true); this.error.set(null); this.success.set(null);
    this.service.createCabinPrice(this.cabinId, {
      ...(value.seasonId ? { seasonId: Number(value.seasonId) } : {}),
      startDate: value.startDate!, endDate: value.endDate!,
      adjustmentPercent: value.pricingMode === 'percent' ? Number(value.adjustmentPercent) : 0,
      ...(value.pricingMode === 'fixed' ? { fixedPrice: Number(value.fixedPrice) } : {}),
    }).subscribe({
      next: () => {
        this.submitting.set(false); this.attempted.set(false);
        this.form.reset({ seasonId: '', startDate: '', endDate: '', pricingMode: 'percent', adjustmentPercent: 0, fixedPrice: null });
        this.load(); this.success.set('Tarifa guardada. Se aplicará a las nuevas reservaciones de esas fechas.');
      },
      error: (err) => { this.submitting.set(false); this.error.set(err.error?.message || 'No pudimos guardar esta tarifa. Inténtalo de nuevo.'); },
    });
  }

  remove(price: CabinSeasonPrice): void {
    if (this.deleting() !== null || this.pendingDelete() !== price.id) return;
    this.deleting.set(price.id); this.error.set(null); this.success.set(null);
    this.service.deleteCabinPrice(price.id).subscribe({
      next: () => { this.prices.update((prices) => prices.filter((p) => p.id !== price.id)); this.pendingDelete.set(null); this.deleting.set(null); this.success.set('Tarifa eliminada. Las reservaciones existentes conservan su total.'); },
      error: (err) => { this.deleting.set(null); this.error.set(err.error?.message || 'No pudimos eliminar esta tarifa.'); },
    });
  }

  priceFor(price: CabinSeasonPrice): number {
    return price.fixedPrice != null ? price.fixedPrice : this.round(this.basePrice * (1 + price.adjustmentPercent / 100));
  }

  labelFor(price: CabinSeasonPrice): string {
    return this.seasons().find((season) => season.id === price.seasonId)?.name || 'Temporada personalizada';
  }

  percent(value: number): string { return `${value > 0 ? '+' : ''}${value}%`; }
  date(value: string): string { return new Date(`${value}T00:00:00`).toLocaleDateString('es-GT', { day: '2-digit', month: 'short', year: 'numeric' }); }
  private round(value: number): number { return Math.round(value * 100) / 100; }
}
