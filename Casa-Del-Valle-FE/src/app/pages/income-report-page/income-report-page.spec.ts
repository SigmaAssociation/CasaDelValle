import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { Cabin } from '../../models/cabin';
import { IncomeReportResponse } from '../../models/report-income';
import { AuthService } from '../../services/auth.service';
import { CabinService } from '../../services/cabin.service';
import { ReportIncomeService } from '../../services/reports-income.service';
import { IncomeReportPage } from './income-report-page';

const cabin: Cabin = {
  id: 1,
  name: 'Cabaña Luna',
  address: 'Km 10',
  price: 500,
  description: 'Cabaña',
  capacity: 4,
  rules: 'Sin mascotas',
};

const report: IncomeReportResponse = {
  movements: [
    {
      id: 1,
      reservation_id: 10,
      cabin_id: 1,
      cabin_name: 'Cabaña Luna',
      guest_name: 'Ana Pérez',
      type: 'anticipo',
      method: 'tarjeta',
      amount: 500,
      date: '2026-10-05T10:00:00Z',
    },
  ],
  total_payments: 500,
  total_refunds: 0,
  total_income: 500,
  total_movements: 1,
  start_date: '2026-10-01',
  end_date: '2026-10-31',
};

const emptyReport: IncomeReportResponse = {
  movements: [],
  total_payments: 0,
  total_refunds: 0,
  total_income: 0,
  total_movements: 0,
  start_date: '2026-10-01',
  end_date: '2026-10-31',
  message: 'No se encontraron ingresos en el rango seleccionado.',
};

describe('IncomeReportPage', () => {
  let fixture: ComponentFixture<IncomeReportPage>;
  let component: IncomeReportPage;
  let getIncomeReport: ReturnType<typeof vi.fn>;
  let getCabins: ReturnType<typeof vi.fn>;
  let getCabinsByUserId: ReturnType<typeof vi.fn>;
  let isAdmin: ReturnType<typeof vi.fn>;
  let getCurrentUserId: ReturnType<typeof vi.fn>;

  function createComponent(admin = true): void {
    getIncomeReport = vi.fn().mockReturnValue(of({ ...report }));
    getCabins = vi.fn().mockReturnValue(of([{ ...cabin }]));
    getCabinsByUserId = vi.fn().mockReturnValue(of([{ ...cabin }]));
    isAdmin = vi.fn().mockReturnValue(admin);
    getCurrentUserId = vi.fn().mockReturnValue(7);

    TestBed.configureTestingModule({
      imports: [IncomeReportPage],
      providers: [
        { provide: ReportIncomeService, useValue: { getIncomeReport } },
        { provide: CabinService, useValue: { getCabins, getCabinsByUserId } },
        { provide: AuthService, useValue: { isAdmin, getCurrentUserId } },
      ],
    });

    fixture = TestBed.createComponent(IncomeReportPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  it('should create and load the current month by default', () => {
    createComponent();

    expect(component).toBeTruthy();

    const now = new Date();
    const start = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`;
    const end = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;

    expect(component.filterForm.value.start_date).toBe(start);
    expect(component.filterForm.value.end_date).toBe(end);
    expect(getIncomeReport).toHaveBeenCalledWith({ startDate: start, endDate: end });
  });

  it('loads all cabins for an admin', () => {
    createComponent(true);

    expect(getCabins).toHaveBeenCalled();
    expect(getCabinsByUserId).not.toHaveBeenCalled();
    expect(component.cabins().length).toBe(1);
  });

  it('loads only the host cabins for a host', () => {
    createComponent(false);

    expect(getCabinsByUserId).toHaveBeenCalledWith(7);
    expect(getCabins).not.toHaveBeenCalled();
  });

  it('blocks the search when the date range is invalid', () => {
    createComponent();
    getIncomeReport.mockClear();

    component.filterForm.patchValue({
      start_date: '2026-10-20',
      end_date: '2026-10-10',
    });
    component.search();

    expect(component.filterForm.hasError('invalidRange')).toBe(true);
    expect(getIncomeReport).not.toHaveBeenCalled();
  });

  it('sends the selected cabin and dates to the service', () => {
    createComponent();
    getIncomeReport.mockClear();

    component.filterForm.patchValue({
      start_date: '2026-10-01',
      end_date: '2026-10-31',
      cabin_id: 3,
    });
    component.search();

    expect(getIncomeReport).toHaveBeenCalledWith({
      startDate: '2026-10-01',
      endDate: '2026-10-31',
      cabinId: 3,
    });
  });

  it('shows the empty state with a zero total', () => {
    createComponent();
    getIncomeReport.mockReturnValue(of({ ...emptyReport }));
    component.search();
    fixture.detectChanges();

    expect(component.report()?.total_income).toBe(0);
    expect(fixture.nativeElement.textContent).toContain('Sin registros');
    expect(fixture.nativeElement.textContent).toContain('No se encontraron ingresos');
  });

  it('shows the backend error message when the request fails', () => {
    createComponent();
    getIncomeReport.mockReturnValue(
      throwError(() => ({ error: { message: 'start_date no puede ser posterior a end_date' } })),
    );
    component.search();
    fixture.detectChanges();

    expect(component.error()).toBe('start_date no puede ser posterior a end_date');
    expect(fixture.nativeElement.textContent).toContain('No se pudo cargar el reporte');
  });
});
