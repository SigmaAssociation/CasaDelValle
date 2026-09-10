import { TestBed } from '@angular/core/testing';

import { CabinService } from './cabin.service';

describe('CabinService', () => {
  let service: CabinService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(CabinService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
