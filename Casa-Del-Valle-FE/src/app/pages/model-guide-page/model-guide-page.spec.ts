import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ModelGuidePage } from './model-guide-page';

describe('ModelGuidePage', () => {
  let component: ModelGuidePage;
  let fixture: ComponentFixture<ModelGuidePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ModelGuidePage],
    }).compileComponents();

    fixture = TestBed.createComponent(ModelGuidePage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
