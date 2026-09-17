import { Component, OnInit } from '@angular/core';
import { CabinListPage } from '../cabin-list-page/cabin-list-page';
import { AuthService } from '../../services/auth.service';

@Component({
  imports: [CabinListPage],
  selector: 'app-my-cabins-page',
  styleUrl: './my-cabins-page.css',
  templateUrl: './my-cabins-page.html',
})
export class MyCabinsPage implements OnInit {
  userId: number | null = null;

  constructor(private auth: AuthService) {}

  ngOnInit(): void {
    this.userId = this.auth.getCurrentUserId();
  }
}
