import { Component, inject, signal } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
    imports: [FormsModule],
    selector: 'app-login',
    styleUrl: './login.css',
    templateUrl: './login.html',
})
export class Login {
    private readonly authService = inject(AuthService);
    private readonly router = inject(Router);

    readonly isLoading = signal(false);
    readonly submitted = signal(false);
    readonly errorMessage = signal('');
    readonly successMessage = signal('');

    public onSubmit(form: NgForm): void {
        this.submitted.set(true);
        this.errorMessage.set('');

        if (form.invalid) {
            return;
        }

        this.isLoading.set(true);

        this.authService.login(form.value.email, form.value.password).subscribe({
            next: (response) => {
                this.isLoading.set(false);

                if (!response.token) {
                    this.errorMessage.set('Correo o contraseña incorrectos.');
                    return;
                }

                this.successMessage.set('Sesión iniciada correctamente. ¡Bienvenido!');

                setTimeout(() => {
                    this.router.navigate(['/']);
                }, 1200);
            },
            error: (err: { status?: number }) => {
                this.isLoading.set(false);
                this.errorMessage.set(
                    err.status === 401 || err.status === 400
                        ? 'Correo o contraseña incorrectos.'
                        : 'No se pudo conectar con el servidor. Inténtalo más tarde.',
                );
            },
        });
    }
}