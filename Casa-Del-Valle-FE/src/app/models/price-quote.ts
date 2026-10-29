export interface PriceQuote {
  cabinId: number; baseNightly: number; totalNights: number; totalPrice: number; averageNightly: number;
  breakdown: { startDate: string; endDate: string; nights: number; baseNightly: number; appliedNightly: number; total: number; source: string; seasonName?: string | null }[];
}
