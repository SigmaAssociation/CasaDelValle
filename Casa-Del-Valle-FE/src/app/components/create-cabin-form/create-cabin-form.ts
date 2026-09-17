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
      name: ['', [
        Validators.required,
        Validators.minLength(3),
        Validators.maxLength(150),
        Validators.pattern(/^[a-zA-ZáéíóúÁÉÍÓÚñÑüÜ0-9\s\-\.\'#]+$/)
      ]],
      address: ['', [
        Validators.required,
        Validators.minLength(5),
        Validators.maxLength(255)
      ]],
      price: ['', [
        Validators.required,
        Validators.min(0)
      ]],
      description: ['', [
        Validators.maxLength(500)
      ]],
      capacity: ['', [
        Validators.required,
        Validators.min(1),
        Validators.max(50)
      ]],
      rules: ['', [
        Validators.maxLength(500)
      ]]
    });
    
    this.fixedHostId = this.hostId ?? this.authService.getCurrentUserId();
  }

  isInvalid(controlName: string): boolean {
    const control = this.cabinForm.get(controlName);
    return !!(control && control.invalid && (control.dirty || control.touched));
  }

  create(): void {
    if (this.cabinForm.valid) {
      const newCabin = this.cabinForm.getRawValue() as CabinRequest;
      newCabin.host_id = this.fixedHostId ?? 0;

      this.cabinService.createCabin(newCabin).subscribe({
        next: (response: RegisterResponse) => {
          console.log(response.message ?? response.mensaje, response.cabin_id);
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
    if (error.error && typeof error.error === 'object' && (error.error.message ?? error.error.mensaje)) {
      return error.error.message ?? error.error.mensaje;
    }

    if (typeof error.error === 'string' && error.error.trim().length > 0) {
      return error.error;
    }

    return 'Ocurrió un error inesperado al registrar la cabaña.';
  }
}