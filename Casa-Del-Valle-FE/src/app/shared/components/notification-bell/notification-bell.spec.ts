import { provideHttpClient } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Notification } from '../../../models/notification';
import { NotificationBell } from './notification-bell';

describe('NotificationBell', () => {
  let component: NotificationBell;
  let fixture: ComponentFixture<NotificationBell>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotificationBell],
      providers: [provideHttpClient(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(NotificationBell);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should resolve the destination link by entity type', () => {
    const base: Notification = {
      id: 1,
      user_id: 1,
      tipo: 'reserva_creada',
      mensaje: 'x',
      leida: false,
      fecha_creacion: '2026-01-01T00:00:00Z',
    };

    expect(component.linkFor({ ...base, entidad_tipo: 'reservacion' })).toBe('/mis-reservaciones');
    expect(component.linkFor({ ...base, entidad_tipo: 'cabana', entidad_id: 7 })).toBe('/cabins/7');
    expect(component.linkFor({ ...base, entidad_tipo: 'cabana' })).toBe('/mis-cabanas');
    expect(component.linkFor({ ...base, entidad_tipo: 'usuario' })).toBe('/perfil');
    expect(component.linkFor({ ...base, entidad_tipo: null })).toBe('/notificaciones');
  });
});
