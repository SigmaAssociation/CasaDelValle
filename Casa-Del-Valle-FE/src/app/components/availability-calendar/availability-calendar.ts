import { Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges } from '@angular/core';
import { DateRange } from '../../models/date-range';

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

const MONTH_NAMES = [
  'Enero',
  'Febrero',
  'Marzo',
  'Abril',
  'Mayo',
  'Junio',
  'Julio',
  'Agosto',
  'Septiembre',
  'Octubre',
  'Noviembre',
  'Diciembre'
];

const WEEKDAY_LABELS = ['Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb', 'Dom'];

export interface CalendarRange {
  start: string | null;
  end: string | null;
}

interface CalendarDay {
  date: string;
  dayNumber: number;
  isCurrentMonth: boolean;
  isPast: boolean;
  isOccupied: boolean;
  isStart: boolean;
  isEnd: boolean;
  isInRange: boolean;
  isToday: boolean;
  isDisabled: boolean;
  isCheckoutOption: boolean;
  cssClass: string;
}

@Component({
  selector: 'app-availability-calendar',
  standalone: true,
  templateUrl: './availability-calendar.html',
  styleUrl: './availability-calendar.css'
})
export class AvailabilityCalendar implements OnInit, OnChanges {
  @Input() blockedRanges: DateRange[] = [];
  @Input() startDate: string | null = null;
  @Input() endDate: string | null = null;
  @Input() minDate: string | null = null;

  @Output() rangeChange = new EventEmitter<CalendarRange>();

  readonly weekdayLabels = WEEKDAY_LABELS;

  private viewDate = new Date();
  private selectedStart: string | null = null;
  private selectedEnd: string | null = null;
  private minDateValue = '';

  ngOnInit(): void {
    this.selectedStart = this.normalizeDate(this.startDate);
    this.selectedEnd = this.normalizeDate(this.endDate);
    this.minDateValue = this.normalizeDate(this.minDate) ?? this.toDateString(new Date());
    this.viewDate = this.initialViewDate();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['minDate']) {
      this.minDateValue = this.normalizeDate(this.minDate) ?? this.minDateValue;
    }
    if (changes['startDate']) {
      this.selectedStart = this.normalizeDate(this.startDate);
    }
    if (changes['endDate']) {
      this.selectedEnd = this.normalizeDate(this.endDate);
    }
  }

  get weeks(): CalendarDay[][] {
    return this.buildWeeks();
  }

  get monthLabel(): string {
    return `${MONTH_NAMES[this.viewDate.getMonth()]} ${this.viewDate.getFullYear()}`;
  }

  get canGoPrevious(): boolean {
    const min = this.parseDate(this.minDateValue);
    if (!min) {
      return false;
    }
    const viewMonth = new Date(this.viewDate.getFullYear(), this.viewDate.getMonth(), 1);
    const minMonth = new Date(min.getFullYear(), min.getMonth(), 1);
    return viewMonth.getTime() > minMonth.getTime();
  }

  get selectedStartLabel(): string {
    return this.formatDisplayDate(this.selectedStart);
  }

  get selectedEndLabel(): string {
    return this.formatDisplayDate(this.selectedEnd);
  }

  previousMonth(): void {
    if (!this.canGoPrevious) {
      return;
    }
    this.viewDate = new Date(this.viewDate.getFullYear(), this.viewDate.getMonth() - 1, 1);
  }

  nextMonth(): void {
    this.viewDate = new Date(this.viewDate.getFullYear(), this.viewDate.getMonth() + 1, 1);
  }

  selectDay(day: CalendarDay): void {
    if (day.isDisabled) {
      return;
    }

    if (!this.selectedStart || this.selectedEnd || day.date <= this.selectedStart) {
      this.selectedStart = day.date;
      this.selectedEnd = null;
      this.emitRange();
      return;
    }

    this.selectedEnd = day.date;
    this.emitRange();
  }

  private emitRange(): void {
    this.rangeChange.emit({
      start: this.selectedStart,
      end: this.selectedEnd
    });
  }

  private buildWeeks(): CalendarDay[][] {
    const year = this.viewDate.getFullYear();
    const month = this.viewDate.getMonth();
    const firstOfMonth = new Date(year, month, 1);
    const leadingDays = (firstOfMonth.getDay() + 6) % 7;
    const gridStart = new Date(year, month, 1 - leadingDays);

    const weeks: CalendarDay[][] = [];
    for (let weekIndex = 0; weekIndex < 6; weekIndex++) {
      const week: CalendarDay[] = [];
      for (let dayIndex = 0; dayIndex < 7; dayIndex++) {
        const offset = weekIndex * 7 + dayIndex;
        const date = new Date(
          gridStart.getFullYear(),
          gridStart.getMonth(),
          gridStart.getDate() + offset
        );
        week.push(this.buildDay(date, year, month));
      }
      weeks.push(week);
    }

    return weeks;
  }

  private buildDay(date: Date, viewYear: number, viewMonth: number): CalendarDay {
    const dateString = this.toDateString(date);
    const isCurrentMonth = date.getFullYear() === viewYear && date.getMonth() === viewMonth;
    const isPast = dateString < this.minDateValue;
    const isOccupied = this.isOccupiedDate(dateString);
    const isStart = this.selectedStart === dateString;
    const isEnd = this.selectedEnd === dateString;
    const isInRange = !!this.selectedStart && !!this.selectedEnd
      && this.selectedStart < dateString && dateString < this.selectedEnd;
    const isToday = dateString === this.toDateString(new Date());
    const isDisabled = this.computeDisabled(dateString, isCurrentMonth, isPast, isOccupied);
    const isCheckoutOption = !isDisabled && this.isCheckoutOption(dateString, isOccupied);

    const day: CalendarDay = {
      date: dateString,
      dayNumber: date.getDate(),
      isCurrentMonth,
      isPast,
      isOccupied,
      isStart,
      isEnd,
      isInRange,
      isToday,
      isDisabled,
      isCheckoutOption,
      cssClass: ''
    };

    day.cssClass = this.buildDayClass(day);
    return day;
  }

  private computeDisabled(
    dateString: string,
    isCurrentMonth: boolean,
    isPast: boolean,
    isOccupied: boolean
  ): boolean {
    if (!isCurrentMonth || isPast) {
      return true;
    }

    if (!this.selectedStart || this.selectedEnd) {
      return isOccupied;
    }

    if (dateString <= this.selectedStart) {
      return isOccupied;
    }

    return this.rangeOverlapsBlocked(this.selectedStart, dateString);
  }

  private isCheckoutOption(dateString: string, isOccupied: boolean): boolean {
    if (!this.selectedStart || this.selectedEnd || !isOccupied) {
      return false;
    }
    if (dateString <= this.selectedStart) {
      return false;
    }
    return !this.rangeOverlapsBlocked(this.selectedStart, dateString);
  }

  private buildDayClass(day: CalendarDay): string {
    const base = 'flex h-9 w-9 items-center justify-center rounded-md text-sm transition-colors';

    if (!day.isCurrentMonth) {
      return `${base} cursor-default text-gray-200`;
    }
    if (day.isStart || day.isEnd) {
      return `${base} bg-orange font-semibold text-white shadow-sm`;
    }
    if (day.isInRange) {
      return `${base} bg-cream text-brown`;
    }
    if (day.isCheckoutOption) {
      return `${base} bg-cream/60 text-brown ring-1 ring-[#D7D0C4] hover:bg-orange/10`;
    }
    if (day.isOccupied) {
      return `${base} cursor-not-allowed bg-gray-100 text-gray-400 line-through`;
    }
    if (day.isPast) {
      return `${base} cursor-not-allowed text-gray-300`;
    }
    if (day.isDisabled) {
      return `${base} cursor-not-allowed text-gray-300`;
    }

    const todayRing = day.isToday ? ' ring-2 ring-orange' : '';
    return `${base} text-gray-700 hover:bg-orange/10${todayRing}`;
  }

  private isOccupiedDate(dateString: string): boolean {
    const time = this.parseDate(dateString)?.getTime();
    if (time === undefined) {
      return false;
    }

    return this.blockedRanges.some((range) => {
      const start = this.parseDate(range.start)?.getTime();
      const end = this.parseDate(range.end)?.getTime();
      if (start === undefined || end === undefined) {
        return false;
      }
      return time >= start && time < end;
    });
  }

  private rangeOverlapsBlocked(startDate: string, endDate: string): boolean {
    const start = this.parseDate(startDate)?.getTime();
    const end = this.parseDate(endDate)?.getTime();
    if (start === undefined || end === undefined) {
      return false;
    }

    return this.blockedRanges.some((range) => {
      const blockedStart = this.parseDate(range.start)?.getTime();
      const blockedEnd = this.parseDate(range.end)?.getTime();
      if (blockedStart === undefined || blockedEnd === undefined) {
        return false;
      }
      return start < blockedEnd && end > blockedStart;
    });
  }

  private initialViewDate(): Date {
    const selected = this.parseDate(this.selectedStart);
    if (selected) {
      return new Date(selected.getFullYear(), selected.getMonth(), 1);
    }

    const min = this.parseDate(this.minDateValue);
    if (min) {
      return new Date(min.getFullYear(), min.getMonth(), 1);
    }

    const today = new Date();
    return new Date(today.getFullYear(), today.getMonth(), 1);
  }

  private normalizeDate(value: string | null | undefined): string | null {
    if (!value || !DATE_PATTERN.test(value)) {
      return null;
    }
    return this.parseDate(value) ? value : null;
  }

  private parseDate(value: string | null | undefined): Date | null {
    if (!value || !DATE_PATTERN.test(value)) {
      return null;
    }

    const [year, month, day] = value.split('-').map(Number);
    const date = new Date(year, month - 1, day);
    if (
      date.getFullYear() !== year ||
      date.getMonth() !== month - 1 ||
      date.getDate() !== day
    ) {
      return null;
    }
    return date;
  }

  private toDateString(date: Date): string {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  private formatDisplayDate(value: string | null): string {
    const date = this.parseDate(value);
    if (!date) {
      return '';
    }
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    return `${day}/${month}/${date.getFullYear()}`;
  }
}
