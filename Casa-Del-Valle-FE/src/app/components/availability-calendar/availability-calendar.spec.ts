import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AvailabilityCalendar } from './availability-calendar';

describe('AvailabilityCalendar', () => {
  let component: AvailabilityCalendar;
  let fixture: ComponentFixture<AvailabilityCalendar>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AvailabilityCalendar],
    }).compileComponents();

    fixture = TestBed.createComponent(AvailabilityCalendar);
    component = fixture.componentInstance;
    fixture.componentRef.setInput('minDate', '2099-06-15');
    fixture.componentRef.setInput('blockedRanges', [
      { start: '2099-06-20', end: '2099-06-23' },
    ]);
    fixture.detectChanges();
  });

  function dayButton(date: string): HTMLButtonElement {
    const button = fixture.nativeElement.querySelector(
      `button[data-date="${date}"]`
    ) as HTMLButtonElement | null;
    if (!button) {
      throw new Error(`No se encontró el día ${date}`);
    }
    return button;
  }

  it('deshabilita las fechas pasadas y las noches ocupadas', () => {
    expect(dayButton('2099-06-10').disabled).toBe(true);
    expect(dayButton('2099-06-21').disabled).toBe(true);
    expect(dayButton('2099-06-22').disabled).toBe(true);
  });

  it('deja libre el día de salida de otra reservación', () => {
    expect(dayButton('2099-06-20').disabled).toBe(true);
    expect(dayButton('2099-06-23').disabled).toBe(false);
  });

  it('permite elegir la salida el día en que inicia otra reservación', () => {
    const emitted: Array<{ start: string | null; end: string | null }> = [];
    component.rangeChange.subscribe((range) => emitted.push(range));

    dayButton('2099-06-18').click();
    fixture.detectChanges();
    dayButton('2099-06-20').click();
    fixture.detectChanges();

    expect(emitted[emitted.length - 1]).toEqual({
      start: '2099-06-18',
      end: '2099-06-20',
    });
  });

  it('permite elegir la entrada el día en que termina otra reservación', () => {
    const emitted: Array<{ start: string | null; end: string | null }> = [];
    component.rangeChange.subscribe((range) => emitted.push(range));

    dayButton('2099-06-23').click();

    expect(emitted[emitted.length - 1]).toEqual({
      start: '2099-06-23',
      end: null,
    });
  });

  it('no permite que la salida cruce una noche ocupada', () => {
    dayButton('2099-06-18').click();
    fixture.detectChanges();

    expect(dayButton('2099-06-22').disabled).toBe(true);
  });
});
