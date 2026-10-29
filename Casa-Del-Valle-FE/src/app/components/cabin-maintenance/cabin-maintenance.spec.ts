import { TestBed, ComponentFixture } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';
import { CabinMaintenance } from './cabin-maintenance';
import { MaintenanceService } from '../../services/maintenance.service';

describe('CabinMaintenance', () => {
  let fixture: ComponentFixture<CabinMaintenance>;
  let component: CabinMaintenance;
  let service: { list: ReturnType<typeof vi.fn>; preview: ReturnType<typeof vi.fn>; create: ReturnType<typeof vi.fn>; cancel: ReturnType<typeof vi.fn> };
  beforeEach(async () => {
    service = { list: vi.fn().mockReturnValue(of([])), preview: vi.fn().mockReturnValue(of({ reservations: [{ id: 5, guest_id: 9, guest_name: 'Huésped', start_date: '2099-01-02', end_date: '2099-01-05', status: 'confirmada', total_price: 300 }] })), create: vi.fn().mockReturnValue(of({ id: 1 })), cancel: vi.fn().mockReturnValue(of({ id: 1 })) };
    await TestBed.configureTestingModule({ imports: [CabinMaintenance], providers: [{ provide: MaintenanceService, useValue: service }] }).compileComponents();
    fixture = TestBed.createComponent(CabinMaintenance); component = fixture.componentInstance;
    fixture.componentRef.setInput('cabinId', 1); fixture.detectChanges();
    component.startDate = '2099-01-02'; component.endDate = '2099-01-04'; component.reason = 'Reparación';
  });
  it('notifica al huésped sin enviar decisiones ni importes del anfitrión', () => {
    component.preview(); fixture.detectChanges();
    expect(component.canSubmit()).toBe(true);
    expect(fixture.nativeElement.querySelectorAll('select').length).toBe(0);
    component.create();
    expect(service.create).toHaveBeenCalledWith(1, { start_date: '2099-01-02', end_date: '2099-01-04', reason: 'Reparación' });
    expect(component.notice()).toContain('huéspedes');
  });
  it('exige revisar de nuevo al cambiar las fechas o recibir un conflicto', () => {
    component.preview(); component.invalidate(); expect(component.canSubmit()).toBe(false);
    component.preview(); service.create.mockReturnValue(throwError(() => ({ error: { message: 'La disponibilidad cambió' } })));
    component.create(); expect(component.error()).toBe('La disponibilidad cambió'); expect(component.canSubmit()).toBe(false);
  });
  it('permite bloquear sin reservas y cancelar antes del inicio', () => {
    service.preview.mockReturnValue(of({ reservations: [] })); component.preview(); expect(component.canSubmit()).toBe(true);
    const changed = vi.spyOn(component.changed, 'emit');
    component.cancel({ id: 7, cabin_id: 1, start_date: '2099-01-02', end_date: '2099-01-04', reason: 'Reparación', cancelled_at: null, created_at: '' });
    expect(service.cancel).toHaveBeenCalledWith(1, 7); expect(changed).toHaveBeenCalled();
  });
});
