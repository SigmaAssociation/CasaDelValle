import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { RegistrarCabana } from './registrar-cabana';

describe('RegistrarCabana', () => {
  let component: RegistrarCabana;
  let fixture: ComponentFixture<RegistrarCabana>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RegistrarCabana],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(RegistrarCabana);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
