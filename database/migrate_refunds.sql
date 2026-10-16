BEGIN;

ALTER TABLE pagos
    DROP CONSTRAINT IF EXISTS chk_pago_tipo;

ALTER TABLE pagos
    ADD CONSTRAINT chk_pago_tipo
    CHECK (tipo_pago IN ('anticipo', 'pago', 'reembolso')) NOT VALID;

COMMIT;
