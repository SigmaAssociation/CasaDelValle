import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { CabinListPage } from './cabin-list-page';

describe('CabinListPage', () => {
  let component: CabinListPage;
  let fixture: ComponentFixture<CabinListPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CabinListPage],
      providers: [provideHttpClient(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(CabinListPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('inicia con los filtros ocultos y los alterna al hacer clic en el botón', () => {
    expect(component.filtersVisible).toBe(false);

    const button = fixture.nativeElement.querySelector('button[type="button"]');
    button.click();
    expect(component.filtersVisible).toBe(true);

    button.click();
    expect(component.filtersVisible).toBe(false);
  });

  it('resolveImageUrl devuelve null cuando no hay imagen', () => {
    expect(component.resolveImageUrl(null)).toBeNull();
    expect(component.resolveImageUrl(undefined)).toBeNull();
    expect(component.resolveImageUrl('')).toBeNull();
  });

  it('resolveImageUrl conserva las URLs absolutas y las rutas que inician con /', () => {
    expect(component.resolveImageUrl('https://ejemplo.com/foto.jpg')).toBe('https://ejemplo.com/foto.jpg');
    expect(component.resolveImageUrl('/uploads/foto.jpg')).toBe('/uploads/foto.jpg');
  });

  it('resolveImageUrl resuelve las rutas relativas contra la URL del API', () => {
    expect(component.resolveImageUrl('uploads/foto.jpg')).toBe('/cdv-api/uploads/foto.jpg');
  });
});
