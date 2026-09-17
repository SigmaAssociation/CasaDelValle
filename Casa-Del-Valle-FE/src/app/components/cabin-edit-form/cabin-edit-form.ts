import { Component, Input, Output, EventEmitter, OnChanges, SimpleChanges, ChangeDetectorRef } from '@angular/core';
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
  
  @Output() onEditSuccess = new EventEmitter<void>();
  successMessage: string | null = null;
  errorMessage: string | null = null;

  isSubmitting: boolean = false;
  private addressRegex = /^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$/;

  constructor(
    private formBuilder: FormBuilder,
    private cabinService: CabinService,
    private cd: ChangeDetectorRef
  ) {
    this.cabinForm = this.formBuilder.group({
      id: [null],
      name: [
        '', 
        [Validators.required, Validators.minLength(3), Validators.maxLength(150)]
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
    if (this.cabinForm.invalid) {
      this.cabinForm.markAllAsTouched();
      return;
    }

    this.successMessage = null;
    this.errorMessage = null;
    this.isSubmitting = true;

    const cabin = this.cabinForm.value as Cabin;
    this.cabinService.editCabin(cabin).subscribe({
      next: () => {
        this.isSubmitting = false;
        this.successMessage = '¡Cabaña actualizada correctamente!';
        this.onEditSuccess.emit();
        setTimeout(() => {
          this.successMessage = null;
          this.cd.detectChanges();
        }, 3000);
      },
      error: (err) => {
        this.isSubmitting = false;
        this.errorMessage = err?.error?.message || err?.error || 'Error al actualizar la cabaña.';
      }
    });
  }
}