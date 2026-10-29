import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Season } from '../../models/season';
import { AdminSeasonsPage } from './admin-seasons-page';

describe('AdminSeasonsPage', () => {
  let fixture: ComponentFixture<AdminSeasonsPage>;
  let component: AdminSeasonsPage;
  let http: HttpTestingController;

  const seasons: Season[] = [
    {
      id: 1,
      name: 'Navidad',
      suggestedAdjustmentPercent: 25,
      active: true,
      startDate: '2026-12-20',
      endDate: '2026-12-31',
    },
    {
      id: 2,
      name: 'Verano',
      suggestedAdjustmentPercent: -10,
      active: false,
      startDate: '2027-06-01',
      endDate: '2027-08-31',
    },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AdminSeasonsPage],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    fixture = TestBed.createComponent(AdminSeasonsPage);
    component = fixture.componentInstance;
    http = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    http.expectOne('/cdv-api/seasons?all=true').flush(seasons);
    fixture.detectChanges();
  });

  afterEach(() => http.verify());

  it('renders active and inactive seasons', () => {
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Navidad');
    expect(text).toContain('Verano');
    expect(text).toContain('Inactiva');
    expect(text).toContain('25%');
  });

  it('creates a season and reloads the list', () => {
    component.openCreate();
    component.form.setValue({
      name: 'Semana Santa',
      suggestedAdjustmentPercent: 15,
      startDate: '2027-03-28',
      endDate: '2027-04-04',
      active: true,
    });
    component.save();

    const request = http.expectOne('/cdv-api/seasons');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({
      name: 'Semana Santa',
      suggestedAdjustmentPercent: 15,
      active: true,
      startDate: '2027-03-28',
      endDate: '2027-04-04',
    });
    request.flush({ id: 3 });
    http.expectOne('/cdv-api/seasons?all=true').flush(seasons);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Temporada creada.');
  });

  it('updates a season and reloads the list', () => {
    component.openEdit(seasons[0]);
    component.form.patchValue({ suggestedAdjustmentPercent: 30 });
    component.save();

    const request = http.expectOne('/cdv-api/seasons/1');
    expect(request.request.method).toBe('PUT');
    expect(request.request.body.suggestedAdjustmentPercent).toBe(30);
    request.flush({ message: 'Temporada actualizada' });
    http.expectOne('/cdv-api/seasons?all=true').flush(seasons);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Temporada actualizada.');
  });

  it('deactivates a season after confirmation', () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    component.deactivate(seasons[0]);

    const request = http.expectOne('/cdv-api/seasons/1/deactivate');
    expect(request.request.method).toBe('PATCH');
    request.flush({ message: 'Temporada desactivada' });
    http.expectOne('/cdv-api/seasons?all=true').flush(seasons);
    fixture.detectChanges();

    expect(confirmSpy).toHaveBeenCalled();
    expect(fixture.nativeElement.textContent).toContain('Temporada desactivada.');
    confirmSpy.mockRestore();
  });

  it('does not deactivate when the user cancels', () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);
    component.deactivate(seasons[0]);
    http.expectNone('/cdv-api/seasons/1/deactivate');
    confirmSpy.mockRestore();
  });

  it('shows backend validation errors in the modal', () => {
    component.openCreate();
    component.form.setValue({
      name: 'Navidad',
      suggestedAdjustmentPercent: 25,
      startDate: '2027-12-20',
      endDate: '2027-12-31',
      active: true,
    });
    component.save();

    const request = http.expectOne('/cdv-api/seasons');
    request.flush(
      { message: 'Ya existe una temporada con fechas cruzadas' },
      { status: 409, statusText: 'Conflict' },
    );
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Ya existe una temporada con fechas cruzadas');
    expect(component.modalOpen()).toBe(true);
  });
});
