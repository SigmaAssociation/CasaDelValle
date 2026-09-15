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
});
