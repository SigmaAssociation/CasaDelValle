import { Component } from '@angular/core';
import { CreateUserForm } from '../../components/create-user-form/create-user-form';

@Component({
  imports: [CreateUserForm],
  selector: 'app-registro',
  styleUrl: './registro.css',
  templateUrl: './registro.html',
})
export class Registro {}
