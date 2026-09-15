import { Component } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CabinListPage } from '../cabin-list-page/cabin-list-page';
import { AuthService } from '../../services/auth.service';

@Component({
  imports: [RouterModule, CabinListPage],
  selector: 'app-start-page',
  styleUrl: './start-page.css',
  templateUrl: './start-page.html',
})
export class StartPage {
  constructor(public auth: AuthService) { }
}
