import { Component, Input, OnChanges, SimpleChanges, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { CabinSeasonPrice, Season } from '../../models/season';
import { SeasonService } from '../../services/season.service';

@Component({
  selector: 'app-cabin-season-prices',
  imports: [ReactiveFormsModule],
  templateUrl: './cabin-season-prices.html',
})
export class CabinSeasonPrices implements OnChanges {
  @Input({ required: true }) cabinId!: number;
  seasons = signal<Season[]>([]);
  prices = signal<CabinSeasonPrice[]>([]);
  error = signal<string | null>(null);
  success = signal<string | null>(null);
  submitting = signal(false);
  readonly form;

  constructor(private fb: FormBuilder, private seasonsService: SeasonService) {
    this.form = this.fb.group({
      seasonId: [''], startDate: ['', Validators.required], endDate: ['', Validators.required],
      adjustmentPercent: [0, Validators.required], fixedPrice: [null as number | null, Validators.min(0.01)],
    });
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['cabinId'] && this.cabinId) this.load();
  }

  load(): void {
    this.seasonsService.list().subscribe({ next: seasons => this.seasons.set(seasons), error: () => this.error.set('No fue posible cargar las temporadas.') });
    this.seasonsService.listCabinPrices(this.cabinId).subscribe({ next: prices => this.prices.set(prices), error: () => this.error.set('No fue posible cargar los precios configurados.') });
  }

  selectSeason(): void {
    const season = this.seasons().find(item => item.id === Number(this.form.controls.seasonId.value));
    if (season) this.form.patchValue({ startDate: season.startDate, endDate: season.endDate, adjustmentPercent: season.suggestedAdjustmentPercent });
  }

  save(): void {
    if (this.form.invalid || this.form.controls.endDate.value! < this.form.controls.startDate.value!) { this.form.markAllAsTouched(); return; }
    const value = this.form.getRawValue();
    this.submitting.set(true); this.error.set(null); this.success.set(null);
    this.seasonsService.createCabinPrice(this.cabinId, {
      seasonId: value.seasonId ? Number(value.seasonId) : undefined, startDate: value.startDate!, endDate: value.endDate!,
      adjustmentPercent: Number(value.adjustmentPercent), fixedPrice: value.fixedPrice == null ? undefined : Number(value.fixedPrice),
    }).subscribe({ next: () => { this.success.set('Precio de temporada guardado.'); this.submitting.set(false); this.form.reset({ adjustmentPercent: 0 }); this.load(); }, error: err => { this.error.set(err?.error?.message ?? 'No fue posible guardar el precio.'); this.submitting.set(false); } });
  }

  remove(price: CabinSeasonPrice): void {
    if (!confirm('¿Eliminar este precio de temporada?')) return;
    this.seasonsService.deleteCabinPrice(price.id).subscribe({ next: () => this.load(), error: err => this.error.set(err?.error?.message ?? 'No fue posible eliminar el precio.') });
  }
}
