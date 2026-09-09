import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BackendGuidePage } from './backend-guide-page';

describe('BackendGuidePage', () => {
  let component: BackendGuidePage;
  let fixture: ComponentFixture<BackendGuidePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [BackendGuidePage],
    }).compileComponents();

    fixture = TestBed.createComponent(BackendGuidePage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
