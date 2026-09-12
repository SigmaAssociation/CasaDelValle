import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { GuidePage } from './guide-page';

describe('GuidePage', () => {
  let component: GuidePage;
  let fixture: ComponentFixture<GuidePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [GuidePage],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(GuidePage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
