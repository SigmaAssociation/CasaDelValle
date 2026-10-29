import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';
import { MaintenanceDecision } from './maintenance-decision';
import { MaintenanceService } from '../../services/maintenance.service';

describe('MaintenanceDecision', () => {
  const details = { reservation_id: 5, cabin_id: 1, reason: 'Reparación', start_date: '2099-01-01', end_date: '2099-01-05', nights: 4, total_price: 808.33, refundable: 242.50, blocked_ranges: [{ start: '2099-01-02', end: '2099-01-04' }] };
  let component: MaintenanceDecision;
  let service: { pending: ReturnType<typeof vi.fn>; resolve: ReturnType<typeof vi.fn> };
  beforeEach(async () => {
    service = { pending: vi.fn().mockReturnValue(of(details)), resolve: vi.fn().mockReturnValue(of({message: 'Decisión guardada'})) };
    await TestBed.configureTestingModule({imports: [MaintenanceDecision], providers: [{provide: MaintenanceService, useValue: service}]}).compileComponents();
    const fixture = TestBed.createComponent(MaintenanceDecision); component = fixture.componentInstance;
    fixture.componentRef.setInput('reservationId', 5); fixture.detectChanges();
  });
  it('no decide al cargar y exige la misma duración y fechas disponibles', () => {
    expect(service.resolve).not.toHaveBeenCalled();
    component.rangeChanged({start: '2099-01-01', end: '2099-01-05'}); expect(component.canReschedule()).toBe(false);
    component.rangeChanged({start: '2099-02-01', end: '2099-02-06'}); expect(component.canReschedule()).toBe(false);
    component.rangeChanged({start: '2099-02-01', end: '2099-02-05'}); expect(component.canReschedule()).toBe(true);
    component.option.set('dates'); component.decide();
    expect(service.resolve).toHaveBeenCalledWith(5, {action: 'reprogramar', start_date: '2099-02-01', end_date: '2099-02-05'});
  });
  it('el huésped solicita el reembolso sin poder alterar su importe', () => {
    component.option.set('refund'); component.decide();
    expect(service.resolve).toHaveBeenCalledWith(5, {action: 'reembolsar'});
  });
  it('muestra conflictos sin emitir una decisión exitosa', () => {
    const resolved = vi.spyOn(component.resolved, 'emit');
    service.resolve.mockReturnValue(throwError(() => ({error: {message: 'Fechas ocupadas'}})));
    component.option.set('refund'); component.decide();
    expect(component.error()).toBe('Fechas ocupadas'); expect(resolved).not.toHaveBeenCalled();
  });
});
