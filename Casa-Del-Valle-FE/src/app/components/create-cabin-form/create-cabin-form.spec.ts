import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateCabinForm } from './create-cabin-form';

describe('CreateCabinForm', () => {
  let component: CreateCabinForm;
  let fixture: ComponentFixture<CreateCabinForm>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CreateCabinForm]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CreateCabinForm);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
