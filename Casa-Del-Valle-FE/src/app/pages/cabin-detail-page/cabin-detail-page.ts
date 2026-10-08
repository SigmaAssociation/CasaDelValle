import { ChangeDetectorRef, Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';
import { Cabin } from '../../models/cabin';
import { CabinImage } from '../../models/cabin-image';
import { CabinService } from '../../services/cabin.service';
import { AuthService } from '../../services/auth.service';
import { ReservationForm } from '../../components/reservation-form/reservation-form';

@Component({
  imports: [CommonModule, RouterLink, ReservationForm],
  selector: 'app-cabin-detail-page',
  templateUrl: './cabin-detail-page.html',
})
export class CabinDetailPage implements OnInit {
  @ViewChild('reservationSection') reservationSection?: ElementRef<HTMLElement>;

  cabin: Cabin | null = null;
  images: CabinImage[] = [];
  selectedImage: string | null = null;
  isLoading = true;
  errorMessage: string | null = null;
  isDeleting = false;
  currentUserId: number | null = null;
  isAdmin = false;
  showReservationForm = false;

  constructor(
    private cabinService: CabinService,
    private authService: AuthService,
    private route: ActivatedRoute,
    private router: Router,
    private cd: ChangeDetectorRef,
  ) { }

  ngOnInit(): void {
    this.currentUserId = this.authService.getCurrentUserId();
    this.isAdmin = this.authService.user()?.role === 1;

    const id = Number(this.route.snapshot.paramMap.get('id'));
    if (!Number.isInteger(id) || id <= 0) {
      this.isLoading = false;
      this.errorMessage = 'ID de cabaña inválido.';
      return;
    }

    forkJoin({
      cabin: this.cabinService.getCabinById(id),
      images: this.cabinService.getImagesByCabin(id),
    }).subscribe({
      next: ({ cabin, images }) => {
        this.cabin = cabin;
        this.images = images ?? [];
        this.selectedImage = this.images.length > 0 ? this.images[0].path : null;
        this.isLoading = false;
        this.cd.detectChanges();
      },
      error: () => {
        this.isLoading = false;
        this.errorMessage = 'No se pudo cargar la cabaña. Es posible que no exista o haya sido eliminada.';
        this.cd.detectChanges();
      },
    });
  }

  canEdit(): boolean {
    return !!this.cabin && (this.isAdmin || this.cabin.host_id === this.currentUserId);
  }

  canReserve(): boolean {
    return !!this.cabin && this.cabin.host_id !== this.currentUserId;
  }

  toggleReservationForm(): void {
    this.showReservationForm = !this.showReservationForm;
    if (this.showReservationForm) {
      // Espera a que Angular renderice el formulario antes de desplazarse.
      setTimeout(() => this.reservationSection?.nativeElement.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
      }));
    }
  }

  formatPrice(price: number | string | null | undefined): string {
    const value = Number(price);
    return Number.isFinite(value) ? `Q${value.toFixed(2)}` : 'Q—';
  }

  resolveImageUrl(path: string | null | undefined): string | null {
    if (!path) {
      return null;
    }
    if (/^https?:\/\//.test(path) || path.startsWith('/')) {
      return path;
    }
    return `${this.cabinService.restConstants.getApiURL()}${path}`;
  }

  onImageError(event: Event): void {
    const target = event.target as HTMLImageElement | null;
    if (target) {
      target.style.visibility = 'hidden';
    }
  }

  deleteCabin(): void {
    if (!this.cabin || this.isDeleting) {
      return;
    }
    if (!confirm(`¿Eliminar "${this.cabin.name}"? Esta acción no se puede deshacer.`)) {
      return;
    }
    this.isDeleting = true;
    this.cabinService.deleteCabin(this.cabin.id).subscribe({
      next: () => {
        void this.router.navigate(['/']);
      },
      error: (err) => {
        console.error('Error al eliminar la cabaña:', err);
        this.errorMessage = 'No se pudo eliminar la cabaña.';
        this.isDeleting = false;
        this.cd.detectChanges();
      },
    });
  }
}
