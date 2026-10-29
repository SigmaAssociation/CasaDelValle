import { ChangeDetectorRef, Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';
import { Cabin } from '../../models/cabin';
import { CabinImage } from '../../models/cabin-image';
import { Reservation } from '../../models/reservation';
import { CancelReservationResponse } from '../../models/cancelReservationResponse';
import { CheckinReservationResponse } from '../../models/checkin-reservation-response';
import { CheckoutReservationResponse } from '../../models/checkout-reservation-response';
import { PayAnticipoResponse } from '../../models/pay-anticipo-response';
import { CreatePaymentResponse } from '../../models/create-payment-response';
import { RefundEvent } from '../../models/refund';
import { CabinService } from '../../services/cabin.service';
import { ReservationService } from '../../services/reservation.service';
import { AuthService } from '../../services/auth.service';
import { ReservationForm } from '../../components/reservation-form/reservation-form';
import { CabinMaintenance } from '../../components/cabin-maintenance/cabin-maintenance';
import { ReservationCard } from '../../components/reservation-card/reservation-card';

@Component({
  imports: [CommonModule, RouterLink, ReservationForm, ReservationCard, CabinMaintenance],
  selector: 'app-cabin-detail-page',
  templateUrl: './cabin-detail-page.html',
})
export class CabinDetailPage implements OnInit {
  @ViewChild('maintenanceSection') maintenanceSection?: ElementRef<HTMLElement>;
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
  cabinReservations: Reservation[] = [];
  isLoadingReservations = false;
  reservationsError: string | null = null;
  reservationNotice: string | null = null;
  reservationsMessage: string | null = null;

  constructor(
    private cabinService: CabinService,
    private reservationService: ReservationService,
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
        if (this.canEdit()) {
          this.loadCabinReservations(id);
        }
        this.cd.detectChanges();
      },
      error: () => {
        this.isLoading = false;
        this.errorMessage = 'No se pudo cargar la cabaña. Es posible que no exista o haya sido eliminada.';
        this.cd.detectChanges();
      },
    });
  }

  loadCabinReservations(cabinId: number): void {
    this.isLoadingReservations = true;
    this.reservationsError = null;
    this.reservationService.getReservationsByCabin(cabinId).subscribe({
      next: (list) => {
        this.cabinReservations = list ?? [];
        this.isLoadingReservations = false;
        this.cd.detectChanges();
      },
      error: (err) => {
        console.error('Error al cargar reservaciones de la cabaña:', err);
        this.isLoadingReservations = false;
        this.reservationsError = 'No se pudieron cargar las reservaciones de esta cabaña.';
        this.cd.detectChanges();
      },
    });
  }

  onReservationCancelled(res: CancelReservationResponse): void {
    this.cabinReservations = this.cabinReservations.map((r) =>
      r.id === res.reservation.id ? { ...r, ...res.reservation } : r,
    );
    this.cd.detectChanges();
  }

  onReservationCheckedIn(res: CheckinReservationResponse): void {
    this.reservationNotice = res.message ?? 'Check-in registrado exitosamente.';
    if (this.cabin) {
      this.loadCabinReservations(this.cabin.id);
    } else {
      this.cd.detectChanges();
    }
  }

  onReservationCheckedOut(res: CheckoutReservationResponse): void {
    this.reservationNotice = res.message ?? 'Check-out registrado exitosamente.';
    if (this.cabin) {
      this.loadCabinReservations(this.cabin.id);
    } else {
      this.cd.detectChanges();
    }
  }

  onReservationPaid(res: PayAnticipoResponse): void {
    this.cabinReservations = this.cabinReservations.map((r) =>
      r.id === res.reservation.id ? { ...r, ...res.reservation } : r,
    );
    this.cd.detectChanges();
  }

  onPaymentRegistered(_res: CreatePaymentResponse): void {
    if (this.cabin) {
      this.loadCabinReservations(this.cabin.id);
    }
  }

  onReservationRefunded(event: RefundEvent): void {
    this.reservationsMessage = event.message;
    this.cd.detectChanges();
  }

  canEdit(): boolean {
    return !!this.cabin && (this.isAdmin || this.cabin.host_id === this.currentUserId);
  }

  canMaintain(): boolean {
    return !!this.cabin && this.authService.isHostOrAdmin() && this.cabin.host_id === this.currentUserId;
  }

  openMaintenance(): void {
    this.maintenanceSection?.nativeElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  onMaintenanceChanged(): void {
    this.showReservationForm = false;
    if (this.cabin) this.loadCabinReservations(this.cabin.id);
  }

  canReserve(): boolean {
    return !!this.cabin && this.cabin.host_id !== this.currentUserId;
  }

  toggleReservationForm(): void {
    this.showReservationForm = !this.showReservationForm;
    if (this.showReservationForm) {
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
