package reservations

import (
	"context"
	"time"
)

// RecordCheckout marca la reservación como finalizada y guarda la fecha y hora
// del check-out. La condición del WHERE hace la operación atómica: solo
// actualiza si sigue en curso y con check-in registrado, de modo que un job
// automático o un doble clic concurrente no puedan finalizarla dos veces.
// GREATEST evita violar chk_checkin_checkout si el reloj de la aplicación
// quedó unos milisegundos detrás del de la base de datos.
func (r *Repository) RecordCheckout(ctx context.Context, id int, checkedOutAt time.Time) error {
	query := `
		UPDATE reservaciones
		SET estado = 'finalizada',
		    fecha_checkout = GREATEST($2::timestamptz, fecha_checkin)
		WHERE id = $1
		  AND estado = 'en_curso'
		  AND fecha_checkin IS NOT NULL
	`
	tag, err := r.pool.Exec(ctx, query, id, checkedOutAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReservationNotCheckoutable
	}
	return nil
}
