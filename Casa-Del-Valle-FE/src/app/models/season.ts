export interface Season {
  id: number;
  name: string;
  suggestedAdjustmentPercent: number;
  active: boolean;
  startDate: string;
  endDate: string;
}

export interface SeasonRequest {
  name: string;
  suggestedAdjustmentPercent: number;
  active: boolean;
  startDate: string;
  endDate: string;
}

export interface CabinSeasonPrice {
  id: number;
  cabinId: number;
  seasonId?: number;
  startDate: string;
  endDate: string;
  adjustmentPercent: number;
  fixedPrice?: number;
}

export interface CabinSeasonPriceRequest {
  seasonId?: number;
  startDate: string;
  endDate: string;
  adjustmentPercent: number;
  fixedPrice?: number;
}
