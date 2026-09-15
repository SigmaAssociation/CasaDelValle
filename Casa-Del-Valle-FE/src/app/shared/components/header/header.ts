import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../../services/auth.service';

@Component({
  selector: 'app-header',
  imports: [RouterModule],
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
