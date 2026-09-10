import { Component, Input, OnInit, signal } from '@angular/core';
import { FormBuilder, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { UserService } from '../../services/user.service';
import { HttpErrorResponse } from '@angular/common/http';
import { UpdateUserRequest } from '../../models/update-user';

@Component({
  standalone: true,
  selector: 'app-edit-user-form',
  imports: [ReactiveFormsModule, FormsModule],
  templateUrl: './edit-user-form.html',
  styleUrl: './edit-user-form.css',
})
export class EditUserForm implements OnInit {
  @Input() userId?: number;

  userForm!: FormGroup;

  searchId = '';

  isLoading = signal(false);
  isLoaded = signal(false);
  isUpdated = signal(false);
  isError = signal(false);
  errorMessage = signal('');
  currentUserId = signal<number | null>(null);
  userDpi = signal('');

  private nameRegex = /^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$/;
  private phoneRegex = /^[123456789]\d{7}$/;
  private addressRegex = /^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$/;
  private emailRegex = /^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,4}$/i;

  constructor(
    private formBuilder: FormBuilder,
    private userService: UserService
  ) { }

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
      email: ['', [
        Validators.required,
        Validators.maxLength(150),
        Validators.pattern(this.emailRegex)
      ]],
    });

    if (this.userId && this.userId > 0) {
      this.searchId = String(this.userId);
      this.load();
    }
  }

  isInvalid(controlName: string): boolean {
    const control = this.userForm.get(controlName);
    return !!(control && control.invalid && (control.dirty || control.touched));
  }

  load(): void {
    const id = Number(this.searchId);
    if (!Number.isInteger(id) || id <= 0) {
      this.isError.set(true);
      this.errorMessage.set('Ingresa un ID de usuario válido.');
      return;
    }

    this.isLoading.set(true);
    this.isError.set(false);
    this.isUpdated.set(false);

    this.userService.getUserById(id).subscribe({
      next: (user) => {
        this.currentUserId.set(user.id);
        this.userDpi.set(user.dpi);
        this.userForm.patchValue({
          name: user.name,
          phone: user.phone,
          address: user.address,
          email: user.email,
        });
        this.userForm.markAsPristine();
        this.userForm.markAsUntouched();
        this.isLoaded.set(true);
        this.isLoading.set(false);
      },
      error: (error: HttpErrorResponse) => {
        this.isLoaded.set(false);
        this.currentUserId.set(null);
        this.isLoading.set(false);
        this.isError.set(true);
        this.errorMessage.set(this.extractErrorMessage(error));
      }
    });
  }

  update(): void {
    const id = this.currentUserId();
    if (id == null) {
      this.isError.set(true);
      this.errorMessage.set('Primero busca y carga un usuario.');
      return;
    }
    if (this.userForm.valid) {
      const payload = this.userForm.value as UpdateUserRequest;

      this.isLoading.set(true);
      this.userService.updateUser(id, payload).subscribe({
        next: () => {
          this.isLoading.set(false);
          this.isUpdated.set(true);
        },
        error: (error: HttpErrorResponse) => {
          this.isLoading.set(false);
          this.isError.set(true);
          this.errorMessage.set(this.extractErrorMessage(error));
        }
      });
    } else {
      this.userForm.markAllAsTouched();
    }
  }

  private extractErrorMessage(error: HttpErrorResponse): string {
    if (error.error && typeof error.error === 'object') {
      if (error.error.message) return error.error.message;
    }

    if (typeof error.error === 'string' && error.error.trim().length > 0) {
      return error.error;
    }

    if (error.status === 404) return 'Usuario no encontrado.';
    if (error.status === 409) return 'El correo ya está registrado por otro usuario.';

    return 'Ocurrió un error inesperado. Intenta de nuevo.';
  }
}
