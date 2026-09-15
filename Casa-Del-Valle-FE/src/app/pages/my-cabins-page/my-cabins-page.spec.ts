import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { MyCabinsPage } from './my-cabins-page';

describe('MyCabinsPage', () => {
  let component: MyCabinsPage;
  let fixture: ComponentFixture<MyCabinsPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MyCabinsPage],
      providers: [provideHttpClient(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(MyCabinsPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
