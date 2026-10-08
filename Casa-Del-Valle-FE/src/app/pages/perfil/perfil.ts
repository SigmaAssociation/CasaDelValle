import { ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { EditUserForm } from '../../components/edit-user-form/edit-user-form';
import { AuthService } from '../../services/auth.service';
import { UserService } from '../../services/user.service';
import { User } from '../../models/user';

@Component({
  imports: [CommonModule, RouterModule, EditUserForm],
  selector: 'app-perfil',
  styleUrl: './perfil.css',
  templateUrl: './perfil.html',
})
export class Perfil implements OnInit {
  userId: number | null = null;
  user: User | null = null;
  isLoading = false;
  editMode = false;
  errorMessage: string | null = null;
  isUpgrading = false;
  upgradeMessage: string | null = null;
  upgradeError: string | null = null;

  private readonly roleNames: Record<number, string> = {
    1: 'Administrador',
    2: 'Huésped',
    3: 'Anfitrión',
  };

  constructor(
    private auth: AuthService,
    private userService: UserService,
    private cd: ChangeDetectorRef
  ) { }

  ngOnInit(): void {
    this.userId = this.auth.getCurrentUserId();
    if (this.userId) {
      this.loadUser(this.userId);
    } else {
      this.errorMessage = 'No se pudo identificar al usuario actual.';
    }
  }

  private loadUser(id: number): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.userService.getUserById(id).subscribe({
      next: (user) => {
        this.user = user;
        this.isLoading = false;
        this.cd.detectChanges();
      },
      error: () => {
        this.errorMessage = 'No se pudo cargar la información del perfil.';
        this.isLoading = false;
        this.cd.detectChanges();
      },
    });
  }

  roleName(role: number | undefined): string {
    if (role === undefined) {
      return '—';
    }
    return this.roleNames[role] ?? '—';
  }

  get canUpgradeToHost(): boolean {
    return this.user?.id_role === 2;
  }

  upgradeToHost(): void {
    if (!this.userId || this.isUpgrading) {
      return;
    }
    if (!confirm('¿Quieres convertirte en anfitrión? Podrás publicar cabañas. Este cambio no se puede deshacer.')) {
      return;
    }
    this.isUpgrading = true;
    this.upgradeMessage = null;
    this.upgradeError = null;

    this.userService.upgradeToHost(this.userId).subscribe({
      next: (response) => {
        if (response.token) {
          this.auth.setToken(response.token);
        }
        this.upgradeMessage = response.message ?? 'Ahora eres anfitrión.';
        this.isUpgrading = false;
        this.loadUser(this.userId as number);
      },
      error: (err) => {
        const backendMessage = err?.error?.message ?? err?.error?.mensaje;
        this.upgradeError = typeof backendMessage === 'string' && backendMessage.trim().length > 0
          ? backendMessage
          : 'No se pudo completar el cambio de rol.';
        this.isUpgrading = false;
        this.cd.detectChanges();
      },
    });
  }
}
