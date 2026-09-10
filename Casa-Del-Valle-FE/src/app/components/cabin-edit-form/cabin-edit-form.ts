import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
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

  successMessage: string | null = null;
  errorMessage: string | null = null;

  private addressRegex = /^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$/;

  constructor(
    private formBuilder: FormBuilder,
    private cabinService: CabinService
  ) {
    this.cabinForm = this.formBuilder.group({
      id: [null],
      address: [
        '',
        [
          Validators.required,
          Validators.maxLength(255),
          Validators.pattern(this.addressRegex)
        ]
      ],
      price: [null, [Validators.required, Validators.min(0)]],
      description: [''],
      capacity: [null, [Validators.required, Validators.min(1)]],
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

    const cabin = this.cabinForm.value as Cabin;

    this.cabinService.editCabin(cabin).subscribe({
      next: () => {
        this.successMessage = '¡Cabaña actualizada correctamente!';
      },
      error: (err) => {
        this.errorMessage = err?.error?.message || err?.error || 'Error al actualizar la cabaña.';
      }
    });
  }
}