import { ChangeDetectorRef, Component, Input, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Cabin } from '../../models/cabin';
import { CabinService } from '../../services/cabin.service';
import { RouterLink } from '@angular/router';

@Component({
  imports: [CommonModule, RouterLink],
  selector: 'app-cabin-list-page',
  styleUrl: './cabin-list-page.css',
  templateUrl: './cabin-list-page.html',
})
export class CabinListPage implements OnInit {
  @Input() userId?: number | null;
  @Input() title = 'Cabañas';
  @Input() subtitle = 'Listado de todas las cabañas registradas.';

  cabins: Cabin[] = [];
  isLoading = false;
  errorMessage: string | null = null;

  constructor(private cabinService: CabinService, private cd : ChangeDetectorRef) { }

  ngOnInit(): void {
    if (this.userId) {
      this.loadByUser(this.userId);
    } else {
      this.loadAll();
    }
  }

  loadAll(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.cabinService.getCabins().subscribe({
      next: (cabins) => {
        this.cabins = cabins;
        this.isLoading = false;
        this.cd.detectChanges();
      },
      error: (err) => {
        console.error('Error al obtener las cabañas:', err);
        this.errorMessage = 'No se pudieron cargar las cabañas.';
        this.isLoading = false;
        this.cd.detectChanges();
      },
    });
  }

  loadByUser(userId: number): void {
    if (!userId || userId <= 0) {
      this.loadAll();
      return;
    }

    this.isLoading = true;
    this.errorMessage = null;

    this.cabinService.getCabinsByUserId(userId).subscribe({
      next: (cabins) => {
        this.cabins = cabins;
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Error al obtener las cabañas del usuario:', err);
        this.errorMessage = 'No se pudieron cargar las cabañas del usuario.';
        this.isLoading = false;
      },
    });
  }
}
