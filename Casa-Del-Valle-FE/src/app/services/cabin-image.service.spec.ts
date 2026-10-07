import { TestBed } from '@angular/core/testing';

import { CabinImageService } from './cabin-image.service';

describe('CabinImageService', () => {
  let service: CabinImageService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(CabinImageService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});