import { ChangeDetectorRef, Component, Input, OnChanges, OnInit, SimpleChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Cabin } from '../../models/cabin';
import { CabinService } from '../../services/cabin.service';
import { RouterLink } from '@angular/router';

@Component({
  imports: [CommonModule, RouterLink],
  selector: 'app-cabin-list-page',
  styleUrl: './cabin-list-page.css',
  templateUrl: './cabin-list-page.html',
})
export class CabinListPage implements OnInit, OnChanges {
  @Input() userId?: number | null;
  @Input() title = 'Cabañas';
  @Input() subtitle = 'Listado de todas las cabañas registradas.';

  cabins: Cabin[] = [];
  isLoading = false;
  errorMessage: string | null = null;
  deletingId: number | null = null;

  constructor(private cabinService: CabinService, private cd: ChangeDetectorRef) { }

  ngOnInit(): void {
    this.load();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['userId'] && !changes['userId'].firstChange) {
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

  trackById(_index: number, cabin: Cabin): number {
    return cabin.id;
  }

  formatPrice(price: number | string | null | undefined): string {
    const value = Number(price);
    return Number.isFinite(value) ? `Q${value.toFixed(2)}` : 'Q—';
  }

  loadAll(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.cabinService.getCabins().subscribe({
      next: (cabins) => {
        this.cabins = cabins ?? [];
        this.isLoading = false;
        this.cd.detectChanges();
      },
      error: (err) => {
        console.error('Error al obtener las cabañas:', err);
        this.errorMessage = this.resolveListError(err, 'No se pudieron cargar las cabañas.');
        this.isLoading = false;
        this.cd.detectChanges();
      },
    });
  }

  loadByUser(userId: number): void {
    if (!Number.isInteger(userId) || userId <= 0) {
      this.cabins = [];
      this.errorMessage = 'ID de usuario inválido.';
      return;
    }

    this.isLoading = true;
    this.errorMessage = null;

    this.cabinService.getCabinsByUserId(userId).subscribe({
      next: (cabins) => {
        this.cabins = cabins ?? [];
        this.isLoading = false;
        this.cd.detectChanges();
      },
      error: (err) => {
        console.error('Error al obtener las cabañas del usuario:', err);
        this.errorMessage = this.resolveListError(err, 'No se pudieron cargar las cabañas del usuario.');
        this.isLoading = false;
        this.cd.detectChanges();
      },
    });
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
