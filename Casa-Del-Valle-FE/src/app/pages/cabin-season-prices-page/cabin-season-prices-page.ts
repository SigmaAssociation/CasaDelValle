import { Component, OnInit, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { CommonModule } from '@angular/common';
import { CabinService } from '../../services/cabin.service';
import { Cabin } from '../../models/cabin';
import { CabinSeasonPrices } from '../../components/cabin-season-prices/cabin-season-prices';

@Component({
  selector: 'app-cabin-season-prices-page',
  standalone: true,
  imports: [CommonModule, RouterLink, CabinSeasonPrices],
  templateUrl: './cabin-season-prices-page.html',
})
export class CabinSeasonPricesPage implements OnInit {
  private readonly service = inject(CabinService);
  private readonly route = inject(ActivatedRoute);
  readonly cabin = signal<Cabin | null>(null);
  readonly loading = signal(true);
  readonly error = signal('');

  ngOnInit(): void {
    const id = Number(this.route.snapshot.paramMap.get('id'));
    if (!Number.isInteger(id) || id <= 0) { this.loading.set(false); this.error.set('La cabaña no es válida.'); return; }
    this.service.getCabinById(id).subscribe({
      next: (cabin) => { this.cabin.set(cabin); this.loading.set(false); },
      error: () => { this.error.set('No se pudo cargar la cabaña. Vuelve al listado e inténtalo de nuevo.'); this.loading.set(false); },
    });
  }
}
