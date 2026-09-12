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
}
