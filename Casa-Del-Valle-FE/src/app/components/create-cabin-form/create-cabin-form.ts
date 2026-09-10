import { HttpErrorResponse } from '@angular/common/http';
import { Component, Input, OnInit, signal } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { CabinRequest } from '../../models/create-cabin';
import { RegisterResponse } from '../../models/register-response';
import { AuthService } from '../../services/auth.service';
import { CabinService } from '../../services/cabin.service';

@Component({
  selector: 'app-create-cabin-form',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './create-cabin-form.html',
  styleUrl: './create-cabin-form.css'
})
export class CreateCabinForm implements OnInit {
  @Input() hostId?: number | null;

  cabinForm!: FormGroup;

  isCreated = signal(false);
  isError = signal(false);
  errorMessage = signal('');

  fixedHostId: number | null = null;

  constructor(
    private formBuilder: FormBuilder,
    private cabinService: CabinService,
    private authService: AuthService
  ) { }

  ngOnInit(): void {
    this.cabinForm = this.formBuilder.group({
      direccion: ['', [
        Validators.required,
        Validators.minLength(5),
        Validators.maxLength(255)
      ]],
      precio: ['', [
        Validators.required,
        Validators.min(0)
      ]],
      descripcion: ['', [
        Validators.maxLength(500)
      ]],
      capacidad: ['', [
        Validators.required,
        Validators.min(1),
        Validators.max(50)
      ]],
      reglas: ['', [
        Validators.maxLength(500)
      ]],
      id_anfitrion: ['', [
        Validators.required,
        Validators.min(1)
      ]],
      id_comision: [null]
    });
    
    this.fixedHostId = this.hostId ?? this.authService.getCurrentUserId();
    if (this.fixedHostId) {
      this.cabinForm.patchValue({ id_anfitrion: this.fixedHostId });
      this.cabinForm.get('id_anfitrion')?.disable();
    }
  }

  isInvalid(controlName: string): boolean {
    const control = this.cabinForm.get(controlName);
    return !!(control && control.invalid && (control.dirty || control.touched));
  }

  create(): void {
    if (this.cabinForm.valid) {
      const newCabin = this.cabinForm.getRawValue() as CabinRequest;

      this.cabinService.createCabin(newCabin).subscribe({
        next: (response: RegisterResponse) => {
          console.log(response.mensaje, response.cabin_id);
          this.isCreated.set(true);
          this.isError.set(false);
        },
        error: (error: HttpErrorResponse) => {
          this.isError.set(true);
          this.errorMessage.set(this.extractErrorMessage(error));
        }
      });
    } else {
      this.cabinForm.markAllAsTouched();
    }
  }

  private extractErrorMessage(error: HttpErrorResponse): string {
    if (error.error && typeof error.error === 'object' && error.error.mensaje) {
      return error.error.mensaje;
    }

    if (typeof error.error === 'string' && error.error.trim().length > 0) {
      return error.error;
    }

    return 'Ocurrió un error inesperado al registrar la cabaña.';
  }
}