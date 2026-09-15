import { Component } from '@angular/core';
import { LoginForm } from '../../components/login-form/login-form';

@Component({
  imports: [LoginForm],
  selector: 'app-login',
  styleUrl: './login.css',
  templateUrl: './login.html',
})
export class Login {}
