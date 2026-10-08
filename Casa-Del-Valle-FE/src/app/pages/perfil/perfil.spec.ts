import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { Perfil } from './perfil';
import { AuthService } from '../../services/auth.service';
import { UserService } from '../../services/user.service';

const user = {
  id: 1,
  name: 'Ana Pérez',
  phone: '5555-5555',
  address: 'Zona 1',
  dpi: '1234567890123',
  email: 'ana@ejemplo.com',
  id_role: 2,
};

describe('Perfil', () => {
  let component: Perfil;
  let fixture: ComponentFixture<Perfil>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Perfil],
      providers: [
        provideHttpClient(),
        provideRouter([]),
        {
          provide: AuthService,
          useValue: {
            getCurrentUserId: () => 1,
            setToken: () => undefined,
          },
        },
        {
          provide: UserService,
          useValue: {
            getUserById: () => of(user),
            upgradeToHost: () => of({ message: 'Ahora eres anfitrión.', token: '' }),
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(Perfil);
    component = fixture.componentInstance;
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
