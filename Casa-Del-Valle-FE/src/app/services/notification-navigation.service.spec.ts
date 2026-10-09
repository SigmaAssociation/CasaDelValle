import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { firstValueFrom } from 'rxjs';
import { Notification } from '../models/notification';
import { NotificationNavigationService } from './notification-navigation.service';

describe('NotificationNavigationService', () => {
  let service: NotificationNavigationService;
  let httpMock: HttpTestingController;

  const base: Notification = {
    id: 1,
    user_id: 1,
    tipo: 'reserva_creada',
    mensaje: 'x',
    leida: false,
    fecha_creacion: '2026-01-01T00:00:00Z',
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(NotificationNavigationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should resolve cabin notifications to the cabin detail view', async () => {
    const link = await firstValueFrom(
      service.resolveLink({ ...base, entidad_tipo: 'cabana', entidad_id: 7 })
    );
    expect(link).toBe('/cabins/7');
  });

  it('should resolve reservation notifications to their cabin detail view', async () => {
    const pending = firstValueFrom(
      service.resolveLink({ ...base, entidad_tipo: 'reservacion', entidad_id: 10 })
    );

    const req = httpMock.expectOne('/cdv-api/reservations/10');
    expect(req.request.method).toBe('GET');
    req.flush({
      id: 10,
      user_id: 1,
      cabin_id: 5,
      host_id: 2,
      start_date: '2026-11-10',
      end_date: '2026-11-12',
      status: 'activa',
    });

    expect(await pending).toBe('/cabins/5');
  });

  it('should fall back to my reservations when the reservation cannot be resolved', async () => {
    const pending = firstValueFrom(
      service.resolveLink({ ...base, entidad_tipo: 'reservacion', entidad_id: 999 })
    );

    const req = httpMock.expectOne('/cdv-api/reservations/999');
    req.flush('No encontrada', { status: 404, statusText: 'Not Found' });

    expect(await pending).toBe('/mis-reservaciones');
  });
});
