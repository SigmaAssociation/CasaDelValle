import { DateRange } from './date-range';

export interface CabinAvailability {
  cabinId: number;
  blockedRanges: DateRange[];
}
