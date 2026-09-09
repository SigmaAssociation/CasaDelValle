import { Component } from '@angular/core';
import { ExampleForm } from '../../components/example-form/example-form';
import { CreateUserForm } from '../../components/create-user-form/create-user-form';

@Component({
  imports: [CreateUserForm],
  selector: 'app-start-page',
  styleUrl: './start-page.css',
  templateUrl: './start-page.html',
})
export class StartPage {}
