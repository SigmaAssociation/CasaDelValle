import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';
import { CabinSeasonPrices } from './cabin-season-prices';
import { SeasonService } from '../../services/season.service';

describe('CabinSeasonPrices', () => {
  const season = { id: 1, name: 'Navidad', suggestedAdjustmentPercent: 25, active: true, startDate: '2099-12-20', endDate: '2099-12-31' };
  let component: CabinSeasonPrices;
  let fixture: ReturnType<typeof TestBed.createComponent<CabinSeasonPrices>>;
  let service: { list: ReturnType<typeof vi.fn>; listCabinPrices: ReturnType<typeof vi.fn>; createCabinPrice: ReturnType<typeof vi.fn>; deleteCabinPrice: ReturnType<typeof vi.fn> };
  beforeEach(async () => {
    service = { list: vi.fn().mockReturnValue(of([season])), listCabinPrices: vi.fn().mockReturnValue(of([])), createCabinPrice: vi.fn().mockReturnValue(of({id: 7})), deleteCabinPrice: vi.fn().mockReturnValue(of({message: 'Eliminado'})) };
    await TestBed.configureTestingModule({imports: [CabinSeasonPrices], providers: [{provide: SeasonService, useValue: service}]}).compileComponents();
    fixture = TestBed.createComponent(CabinSeasonPrices); component = fixture.componentInstance;
    fixture.componentRef.setInput('cabinId', 3); fixture.componentRef.setInput('basePrice', 400); fixture.detectChanges();
  });
  function chooseSeason() { component.form.controls.seasonId.setValue('1'); component.selectSeason(); }

  it('muestra el precio recomendado y permite recuperar la sugerencia después de personalizarla', () => {
    chooseSeason(); fixture.detectChanges();
    expect(component.recommendedPrice).toBe(500);
    expect(component.previewPrice).toBe(500);
    expect(component.form.controls.startDate.value).toBe('2099-12-20');
    expect(fixture.nativeElement.textContent).toContain('Precio recomendado');
    expect(fixture.nativeElement.textContent).toContain('Q500.00');
    component.form.patchValue({pricingMode: 'fixed', fixedPrice: 650});
    expect(component.previewPrice).toBe(650);
    expect(component.recommendedPrice).toBe(500);
    component.useRecommendation();
    expect(component.form.controls.fixedPrice.value).toBeNull();
    expect(component.previewPrice).toBe(500);
  });

  it('envía solo el precio fijo activo sin acumular el porcentaje recomendado', () => {
    chooseSeason(); component.form.patchValue({pricingMode: 'fixed', fixedPrice: 650}); component.save();
    expect(service.createCabinPrice).toHaveBeenCalledWith(3, {seasonId: 1, startDate: '2099-12-20', endDate: '2099-12-31', adjustmentPercent: 0, fixedPrice: 650});
  });

  it('permite descuentos y elimina del envío un precio fijo anterior al elegir porcentaje', () => {
    chooseSeason(); component.form.patchValue({pricingMode: 'percent', adjustmentPercent: -10, fixedPrice: 650});
    expect(component.previewPrice).toBe(360);
    component.save();
    expect(service.createCabinPrice).toHaveBeenCalledWith(3, {seasonId: 1, startDate: '2099-12-20', endDate: '2099-12-31', adjustmentPercent: -10});
  });

  it('rechaza fechas invertidas y precios inválidos; acepta una sola noche', () => {
    chooseSeason(); component.form.patchValue({endDate: '2099-12-19'}); component.save();
    expect(component.form.hasError('dateOrder')).toBe(true); expect(service.createCabinPrice).not.toHaveBeenCalled();
    component.form.patchValue({endDate: '2099-12-20', adjustmentPercent: -100}); component.save();
    expect(component.form.hasError('invalidPercent')).toBe(true); expect(service.createCabinPrice).not.toHaveBeenCalled();
    component.form.patchValue({adjustmentPercent: 0}); component.save(); expect(service.createCabinPrice).toHaveBeenCalledTimes(1);
  });

  it('conserva los datos al mostrar un conflicto de temporadas', () => {
    chooseSeason(); service.createCabinPrice.mockReturnValue(throwError(() => ({error: {message: 'Ya existe un precio para ese rango de fechas'}})));
    component.save(); expect(component.error()).toContain('Ya existe'); expect(component.form.controls.seasonId.value).toBe('1'); expect(component.submitting()).toBe(false);
  });
});
