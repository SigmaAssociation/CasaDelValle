import { Component, OnInit, signal } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, ReactiveFormsModule, ValidationErrors, ValidatorFn, Validators } from '@angular/forms';
import { UserRequest } from '../../models/create-user';
import { UserService } from '../../services/user.service';
import { HttpErrorResponse } from '@angular/common/http';
import { RegisterResponse } from '../../models/register-response';

@Component({
  standalone: true,
  selector: 'app-create-user-form',
  imports: [ReactiveFormsModule],
  templateUrl: './create-user-form.html',
  styleUrl: './create-user-form.css',
})
export class CreateUserForm implements OnInit {
  userForm!: FormGroup;

  isCreated = signal(false);
  isError = signal(false);
  errorMessage = signal('');

  private nameRegex = /^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$/;
  private phoneRegex = /^[123456789]\d{7}$/;
  private addressRegex = /^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$/;
  private dpiRegex = /^\d{13}$/;
  private emailRegex = /^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,4}$/i;

  constructor(
    private formBuilder: FormBuilder,
    private userService: UserService
  ) { }


  private passwordStrengthValidator(): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      const value: string = control.value || '';

      const hasUpperCase = /[A-Z]/.test(value);
      const hasLowerCase = /[a-z]/.test(value);
      const hasDigit = /\d/.test(value);
      const hasSpecialChar = /[!@#~$%^&*()+|_.,<>?/\\-]/.test(value);
      const hasMinLength = value.length >= 8;

      const errors = { hasUpperCase, hasLowerCase, hasDigit, hasSpecialChar, hasMinLength };
      const isValid = hasUpperCase && hasLowerCase && hasDigit && hasSpecialChar && hasMinLength;

      return isValid ? null : { passwordStrength: errors };
    };
  }

  ngOnInit(): void {
    this.userForm = this.formBuilder.group({
      name: ['', [
        Validators.required,
        Validators.minLength(2),
        Validators.maxLength(150),
        Validators.pattern(this.nameRegex)
      ]],
      phone: ['', [
        Validators.required,
        Validators.pattern(this.phoneRegex)
      ]],
      address: ['', [
        Validators.required,
        Validators.minLength(5),
        Validators.maxLength(255),
        Validators.pattern(this.addressRegex)
      ]],
      dpi: ['', [
        Validators.required,
        Validators.pattern(this.dpiRegex)
      ]],
      email: ['', [
        Validators.required,
        Validators.maxLength(150),
        Validators.pattern(this.emailRegex)
      ]],
      password: ['', [
        Validators.required,
        Validators.maxLength(255),
        this.passwordStrengthValidator()
      ]]
    });
  }

  isInvalid(controlName: string): boolean {
    const control = this.userForm.get(controlName);
    return !!(control && control.invalid && (control.dirty || control.touched));
  }

  passwordMeets(rule: 'hasUpperCase' | 'hasLowerCase' | 'hasDigit' | 'hasSpecialChar' | 'hasMinLength'): boolean {
    const control = this.userForm.get('password');
    if (!control) return false;
    if (!control.errors?.['passwordStrength']) return control.value?.length > 0;
    return control.errors['passwordStrength'][rule];
  }

  get showPasswordChecklist(): boolean {
    const control = this.userForm.get('password');
    return !!(control && (control.dirty || control.touched) && control.value);
  }

  create(): void {
    if (this.userForm.valid) {
      const newUser = this.userForm.value as UserRequest;

      this.userService.createUser(newUser).subscribe({
        next: (response: RegisterResponse) => {
          console.log(response.mensaje, response.user_id);
          this.isCreated.set(true);
        },
        error: (error: HttpErrorResponse) => {
          this.isError.set(true);
          this.errorMessage.set(this.extractErrorMessage(error));
        }
      });
    } else {
      this.userForm.markAllAsTouched();
    }
  }

  private extractErrorMessage(error: HttpErrorResponse): string {
    if (error.error && typeof error.error === 'object' && error.error.mensaje) {
      return error.error.mensaje;
    }

    if (typeof error.error === 'string' && error.error.trim().length > 0) {
      return error.error;
    }

    return 'Ocurrió un error inesperado. Intenta de nuevo.';
  }
}