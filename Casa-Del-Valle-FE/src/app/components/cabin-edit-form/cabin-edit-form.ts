import { Component, EventEmitter, Input, OnChanges, Output, SimpleChanges, signal } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Cabin } from '../../models/cabin';
import { CabinImage } from '../../models/cabin-image';
import { CabinService } from '../../services/cabin.service';

@Component({
  imports: [ReactiveFormsModule],
  selector: 'app-cabin-edit-form',
  styleUrl: './cabin-edit-form.css',
  templateUrl: './cabin-edit-form.html',
})
export class CabinEditForm implements OnChanges {
  cabinForm!: FormGroup;

  @Input({ required: true })
  cabin!: Cabin;

  @Output()
  onEditSuccess = new EventEmitter<void>();

  successMessage = signal<string | null>(null);
  errorMessage = signal<string | null>(null);
  isSubmitting = signal(false);

  images = signal<CabinImage[]>([]);
  isLoadingImages = signal(false);
  uploadError = signal<string | null>(null);
  isUploading = signal(false);

  private addressRegex = /^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$/;
  private nameRegex = /^[a-zA-ZáéíóúÁÉÍÓÚñÑüÜ0-9\s\-\.\'#]+$/;

  constructor(
    private formBuilder: FormBuilder,
    private cabinService: CabinService
  ) {
    this.cabinForm = this.formBuilder.group({
      id: [null],
      name: [
        '',
        [
          Validators.required,
          Validators.maxLength(150),
          Validators.pattern(this.nameRegex),
          Validators.minLength(3)
        ]
      ],
      address: [
        '',
        [
          Validators.required,
          Validators.minLength(5),
          Validators.maxLength(255),
          Validators.pattern(this.addressRegex)
        ]
      ],
      price: [null, [Validators.required, Validators.min(0.01)]],
      description: ['', [Validators.maxLength(500)]],
      capacity: [null, [Validators.required, Validators.min(1), Validators.max(50)]],
      rules: [''],
    });
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['cabin'] && this.cabin) {
      this.cabinForm.patchValue(this.cabin);
      this.loadImages();
    }
  }

  loadImages(): void {
    if (!this.cabin?.id) return;
    this.isLoadingImages.set(true);
    this.cabinService.getImagesByCabin(this.cabin.id).subscribe({
      next: (images) => {
        this.images.set(images);
        this.isLoadingImages.set(false);
      },
      error: () => {
        this.isLoadingImages.set(false);
      }
    });
  }

  onFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.uploadImage(input.files[0]);
    }
  }

  uploadImage(file: File): void {
    if (!this.cabin?.id) return;
    this.uploadError.set(null);
    this.isUploading.set(true);

    const allowedTypes = ['image/jpeg', 'image/png', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      this.uploadError.set('Formato no válido. Solo JPEG, PNG o WebP.');
      this.isUploading.set(false);
      return;
    }
    const maxSize = 5 * 1024 * 1024;
    if (file.size > maxSize) {
      this.uploadError.set('La imagen excede el tamaño máximo de 5 MB.');
      this.isUploading.set(false);
      return;
    }

    this.cabinService.uploadImage(this.cabin.id, file).subscribe({
      next: (response) => {
        this.isUploading.set(false);
        this.loadImages();
      },
      error: (err) => {
        this.isUploading.set(false);
        const msg = err?.error?.mensaje ?? err?.error?.message ?? 'Error al subir la imagen.';
        this.uploadError.set(msg);
      }
    });
  }

  deleteImage(imageId: number): void {
    if (!confirm('¿Eliminar esta imagen?')) return;
    this.cabinService.deleteImage(imageId).subscribe({
      next: () => this.loadImages(),
      error: (err) => {
        const msg = err?.error?.mensaje ?? err?.error?.message ?? 'Error al eliminar la imagen.';
        this.uploadError.set(msg);
      }
    });
  }

  getImageUrl(ruta: string): string {
    return `${this.cabinService.restConstants.getApiURL()}${ruta}`;
  }

  edit(): void {
    if (this.cabinForm.invalid || this.isSubmitting()) {
      this.cabinForm.markAllAsTouched();
      return;
    }

    this.successMessage.set(null);
    this.errorMessage.set(null);
    this.isSubmitting.set(true);

    const raw = this.cabinForm.value;
    const cabin: Cabin = {
      ...raw,
      price: Number(raw.price),
      capacity: Number(raw.capacity),
    };

    this.cabinService.editCabin(cabin).subscribe({
      next: () => {
        this.successMessage.set('¡Cabaña actualizada correctamente!');
        this.isSubmitting.set(false);
        this.onEditSuccess.emit();
      },
      error: (err) => {
        const backendMessage = err?.error?.message ?? err?.error?.mensaje;
        this.errorMessage.set(
          typeof backendMessage === 'string' && backendMessage.trim().length > 0
            ? backendMessage
            : 'Error al actualizar la cabaña.',
        );
        this.isSubmitting.set(false);
      }
    });
  }
}