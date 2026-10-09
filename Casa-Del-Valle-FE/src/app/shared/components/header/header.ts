import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../../services/auth.service';
import { NotificationBell } from '../notification-bell/notification-bell';

@Component({
  selector: 'app-header',
  imports: [RouterModule, NotificationBell],
  templateUrl: './header.html',
  styleUrl: './header.css'
})
export class Header {
  constructor(
    public auth: AuthService,
    private router: Router
  ) { }

  onLogout(): void {
    this.auth.logout();
    this.router.navigate(['/']);
  }
}
