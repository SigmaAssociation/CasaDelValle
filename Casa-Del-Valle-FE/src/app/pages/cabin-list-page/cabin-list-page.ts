import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Cabin } from '../../models/cabin';
import { CabinService } from '../../services/cabin.service';
import { RouterLink } from '@angular/router';

@Component({
  imports: [CommonModule, RouterLink],
  selector: 'app-cabin-list-page',
  styleUrl: './cabin-list-page.css',
  templateUrl: './cabin-list-page.html',
})
export class CabinListPage implements OnInit {
  cabins: Cabin[] = [];

  constructor(private cabinService: CabinService) { }

  ngOnInit(): void {
    const userId = this.getCurrentUserId();
    this.cabinService.getCabinsByUserId(userId).subscribe({
      next: (cabins) => {
        this.cabins = cabins;
      },
      error: (err) => {
        console.error('Error al obtener las cabañas:', err);
      },
    });
  }

  private getCurrentUserId(): number {
    //para el test
    return 1;
  }
}