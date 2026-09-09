import { ComponentFixture, TestBed } from '@angular/core/testing';
import { DesignGuidePage } from './design-guide-page';

describe('DesignGuidePage', () => {
  let component: DesignGuidePage;
  let fixture: ComponentFixture<DesignGuidePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DesignGuidePage],
    }).compileComponents();

    fixture = TestBed.createComponent(DesignGuidePage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
