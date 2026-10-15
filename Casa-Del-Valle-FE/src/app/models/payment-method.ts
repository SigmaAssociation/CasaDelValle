export type PaymentMethod = 'tarjeta' | 'transferencia' | 'efectivo';

export interface PaymentMethodOption {
  value: PaymentMethod;
  label: string;
}

export const PAYMENT_METHODS: PaymentMethodOption[] = [
  { value: 'tarjeta', label: 'Tarjeta' },
  { value: 'transferencia', label: 'Transferencia' },
  { value: 'efectivo', label: 'Efectivo (día del check-in)' },
];

export function paymentMethodLabel(method: string): string {
  switch (method) {
    case 'tarjeta':
      return 'Tarjeta';
    case 'transferencia':
      return 'Transferencia';
    case 'efectivo':
      return 'Efectivo';
    case 'anticipo':
      return 'Anticipo';
    default:
      return method;
  }
}
