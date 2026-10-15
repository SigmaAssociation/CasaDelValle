import { CurrencyPipe, DatePipe } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ReportReservationsService } from '../../services/reports-reservations.service';
import { ReservationsReportFilter } from '../../models/report-reservations';
import { ReservationsReportResponse } from '../../models/report-reservations';

@Component({
  selector: 'app-reporte-lista-reservaciones',
  standalone: true,
  imports: [FormsModule, CurrencyPipe, DatePipe],
  templateUrl: './reporte-lista-reservaciones.html',
})
export class ReporteListaReservacionesComponent implements OnInit {
  reporte = signal<ReservationsReportResponse | null>(null);
  cargando = signal(false);
  error = signal('');
  mostrarFiltros = signal(false);
  
  filtro: { host_id: number | null; cabin_id: number | null } = { host_id: null, cabin_id: null };

  constructor(private reportService: ReportReservationsService) { }

  ngOnInit(): void {
    this.cargarReporte();
  }

  buscar(): void {
    this.cargarReporte();
  }

  limpiar(): void {
    this.filtro = { host_id: null, cabin_id: null };
    this.cargarReporte();
  }

  etiquetaEstado(estado: string): string {
    return estado.replace('_', ' ');
  }

  claseEstado(estado: string): string {
    const clases: Record<string, string> = {
      pendiente: 'bg-yellow-100 text-yellow-800',
      confirmada: 'bg-green-100 text-green-800',
      en_curso: 'bg-blue-100 text-blue-800',
      finalizada: 'bg-gray-100 text-gray-600',
      cancelada: 'bg-red-100 text-red-800',
    };
    return clases[estado] ?? 'bg-gray-100 text-gray-600';
  }

  private cargarReporte(): void {
    this.cargando.set(true);
    this.error.set('');

    const filter: ReservationsReportFilter = {};
    if (this.filtro.host_id != null) filter.host_id = this.filtro.host_id;
    if (this.filtro.cabin_id != null) filter.cabin_id = this.filtro.cabin_id;

    this.reportService.getReservationsReport(filter).subscribe({
      next: (reporte) => {
        this.reporte.set(reporte);
        this.cargando.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.reporte.set(null);
        this.error.set(err.error?.message ?? 'No se pudo cargar el reporte.');
        this.cargando.set(false);
      },
    });
  }
}