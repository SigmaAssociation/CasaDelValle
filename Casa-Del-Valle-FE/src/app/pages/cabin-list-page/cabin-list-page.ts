import { ChangeDetectorRef, Component, Input, OnChanges, OnInit, SimpleChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { CabinCardResponse, CabinSearchParams } from '../../models/cabin-response';
import { CabinService } from '../../services/cabin.service';
import { AuthService } from '../../services/auth.service';
import { RouterLink } from '@angular/router';

@Component({
  imports: [CommonModule, RouterLink, FormsModule],
  selector: 'app-cabin-list-page',
  styleUrl: './cabin-list-page.css',
  templateUrl: './cabin-list-page.html',
})
export class CabinListPage implements OnInit, OnChanges {
  @Input() userId?: number | null;
  @Input() title = 'Cabañas';
  @Input() subtitle = 'Listado de todas las cabañas registradas.';

  cabins: CabinCardResponse[] = [];
  totalCabins = 0; 
  isLoading = false;
  errorMessage: string | null = null;
  deletingId: number | null = null;
  currentUserId: number | null = null; 
  isAdmin = false;

  filters: CabinSearchParams = { 
    name: '',
    host_name: null,
    host_id: null,
    min_capacity: null,
    max_capacity: null,
    min_price: null,
    max_price: null
  };

  constructor(private cabinService: CabinService, private authService: AuthService, private cd: ChangeDetectorRef) { }

  ngOnInit(): void {
    this.currentUserId = this.authService.getCurrentUserId();
    this.isAdmin = this.authService.user()?.role === 1;
    this.load();
  }

  ngOnChanges(changes: SimpleChanges): void {
    const userIdChange = changes['userId'];
    if (userIdChange && userIdChange.currentValue != null
      && userIdChange.currentValue !== userIdChange.previousValue) {
      this.load();
    }
  }

  load(): void {
    if (this.userId != null) {
      this.loadByUser(this.userId);
    } else {
      this.loadAll();
    }
  }

  trackById(_index: number, cabin: CabinCardResponse): number {
    return cabin.id;
  }

  formatPrice(price: number | string | null | undefined): string {
    const value = Number(price);
    return Number.isFinite(value) ? `Q${value.toFixed(2)}` : 'Q—';
  }

  loadAll(): void {
    this.applyFilters();
  }

  loadByUser(userId: number): void {
    if (!Number.isInteger(userId) || userId <= 0) {
      this.cabins = [];
      this.errorMessage = 'ID de usuario inválido.';
      return;
    }
    this.userId = userId;
    this.filters.host_id = userId;
    this.applyFilters();
  }

  applyFilters(): void {
    this.isLoading = true;
    this.errorMessage = null;
    const searchParams: CabinSearchParams = {
      ...this.filters,
    };

    this.cabinService.searchCabins(searchParams).subscribe({
      next: (response) => {
        this.cabins = response.data || [];
        this.totalCabins = response.total || 0;
        this.isLoading = false;
        this.cd.detectChanges();
      },
      error: (err) => {
        console.error('Error al buscar cabañas:', err);
        this.errorMessage = this.resolveListError(err, 'No se pudieron cargar las cabañas.');
        this.isLoading = false;
        this.cd.detectChanges();
      }
    });
  }

  clearFilters(): void {
    this.filters = { 
      name: '', 
      host_name: null,
      host_id: this.userId ?? null,
      min_price: null, 
      max_price: null, 
      min_capacity: null, 
      max_capacity: null 
    };
    this.applyFilters();
  }

  canEdit(hostId: number): boolean {
    return this.isAdmin || hostId === this.currentUserId;
  }

  deleteCabin(id: number): void {
    if (!Number.isInteger(id) || id <= 0 || this.deletingId !== null) {
      return;
    }

    if (!confirm('¿Eliminar esta cabaña? Esta acción no se puede deshacer.')) {
      return;
    }

    this.deletingId = id;
    this.errorMessage = null;

    this.cabinService.deleteCabin(id).subscribe({
      next: () => {
        this.cabins = this.cabins.filter((cabin) => cabin.id !== id);
        this.deletingId = null;
        this.cd.detectChanges();
      },
      error: (err) => {
        console.error('Error al eliminar la cabaña:', err);
        this.errorMessage = this.resolveListError(err, 'No se pudo eliminar la cabaña.');
        this.deletingId = null;
        this.cd.detectChanges();
      },
    });
  }

  private resolveListError(err: unknown, fallback: string): string {
    if (err instanceof HttpErrorResponse) {
      if (err.status === 0) {
        return 'No se pudo conectar con el servidor. Verifica que el backend esté en ejecución.';
      }
      if (err.status === 401) {
        return 'Sesión no válida o expirada. Inicia sesión nuevamente.';
      }
      if (err.status === 403) {
        return 'No tienes permisos para ver estas cabañas.';
      }
      if (err.status === 404) {
        return 'Recurso no encontrado.';
      }
      const backendMessage = (err.error as { message?: string; mensaje?: string } | null)?.message
        ?? (err.error as { mensaje?: string } | null)?.mensaje;
      if (typeof backendMessage === 'string' && backendMessage.trim().length > 0) {
        return backendMessage;
      }
    }
    return fallback;
  }
}