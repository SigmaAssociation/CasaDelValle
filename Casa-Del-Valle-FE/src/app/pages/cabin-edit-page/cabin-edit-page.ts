import { ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { CabinEditForm } from "../../components/cabin-edit-form/cabin-edit-form";
import { Cabin } from '../../models/cabin';
import { CabinService } from '../../services/cabin.service';

@Component({
  imports: [CabinEditForm, RouterLink],
  selector: 'app-cabin-edit-page',
  styleUrl: './cabin-edit-page.css',
  templateUrl: './cabin-edit-page.html',
})
export class CabinEditPage implements OnInit {
  cabin?: Cabin;
  notFound = false;

  constructor(
    private cabinService: CabinService,
    private route: ActivatedRoute,
    private cd : ChangeDetectorRef,
    private router: Router
  ) { }

  ngOnInit(): void {
    const idParam = this.route.snapshot.paramMap.get('id');
    const id = Number(idParam);

    if (!idParam || isNaN(id)) {
      this.notFound = true;
      return;
    }

    this.cabinService.getCabinById(id).subscribe({
      next: (cabin) => {
        if (cabin) {
          this.cabin = cabin;
          this.notFound = false;
        } else {
          this.notFound = true;
        }
        this.cd.detectChanges();
      },
      error: (err) => {
        console.error('Error al obtener la cabaña:', err);
        this.notFound = true;
        this.cd.detectChanges();
      },
    });
  }
  onEditSuccess(): void {
  }
}