import { Component } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CreateUserForm } from '../../components/create-user-form/create-user-form';
import { EditUserForm } from '../../components/edit-user-form/edit-user-form';

@Component({
  imports: [CreateUserForm, EditUserForm, RouterModule],
  selector: 'app-start-page',
  styleUrl: './start-page.css',
  templateUrl: './start-page.html',
})
export class StartPage {}