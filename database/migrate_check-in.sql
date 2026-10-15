ALTER TABLE notificaciones
ALTER CONSTRAINT chk_notificacion_tipo
    ADD CHECK (tipo IN (
            'bienvenida',
            'reserva_creada',
            'reserva_cancelada',
            'reserva_actualizada',
            'reserva_finalizada',
            'cabana_creada',
            'cabana_actualizada',
            'cabana_eliminada',
            'rol_actualizado',
            'reserva_checkin'
        ));
