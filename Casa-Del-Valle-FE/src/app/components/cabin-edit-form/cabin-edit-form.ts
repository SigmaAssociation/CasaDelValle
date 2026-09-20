import { Component, EventEmitter, Input, OnChanges, Output, SimpleChanges, signal } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Cabin } from '../../models/cabin';
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
    }
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