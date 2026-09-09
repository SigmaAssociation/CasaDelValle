import { ComponentFixture, TestBed } from '@angular/core/testing';
import { UserStoryGuidePage } from './user-story-guide-page';

describe('UserStoryGuidePage', () => {
  let component: UserStoryGuidePage;
  let fixture: ComponentFixture<UserStoryGuidePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UserStoryGuidePage],
    }).compileComponents();

    fixture = TestBed.createComponent(UserStoryGuidePage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
