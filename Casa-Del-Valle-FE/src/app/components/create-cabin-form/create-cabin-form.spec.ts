import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { provideRouter, Router } from '@angular/router';

import { CreateCabinForm } from './create-cabin-form';

describe('CreateCabinForm', () => {
  let component: CreateCabinForm;
  let fixture: ComponentFixture<CreateCabinForm>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CreateCabinForm],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    })
    .compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(CreateCabinForm);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('redirige a mis cabañas después de un registro exitoso', () => {
    const router = TestBed.inject(Router);
    const navigateSpy = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    component.cabinForm.setValue({
      name: 'Cabaña de prueba',
      address: 'Ruta 1, Quetzaltenango',
      price: 100,
      description: '',
      capacity: 4,
      rules: '',
    });
    component.create();

    const req = httpMock.expectOne('/cdv-api/cabins');
    expect(req.request.method).toBe('POST');
    req.flush({ message: 'Cabaña registrada', cabin_id: 1 });

    expect(component.isCreated()).toBe(true);

    component.goToMyCabins();
    expect(navigateSpy).toHaveBeenCalledWith(['/mis-cabanas']);
  });
});
