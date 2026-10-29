package maintenance_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"cdv-api/internal/maintenance"
	"cdv-api/internal/middleware"
	"cdv-api/internal/payments"
	"cdv-api/internal/reservations"
	"cdv-api/internal/seasons"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Cada prueba crea su propio esquema; no elimina datos de la base configurada.
func fixture(t *testing.T, beforeTotal ...func(*pgxpool.Pool)) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("MAINTENANCE_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("configura MAINTENANCE_TEST_DATABASE_URL para las pruebas PostgreSQL")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := "maintenance_test_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	ident := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+ident); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close(); admin.Exec(ctx, "DROP SCHEMA "+ident+" CASCADE"); admin.Close() })
	files := []string{"../../../database/init.sql"}
	migrations, err := filepath.Glob("../../../database/migrations/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, migrations...)
	for _, f := range files {
		if filepath.Base(f) == "009_guest_maintenance_and_total.sql" {
			for _, seed := range beforeTotal {
				seed(pool)
			}
		}
		sql, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		// El job global de pg_cron se verifica con migrate.sh en Docker;
		// aquí solo se crean objetos propios del esquema de la prueba.
		if filepath.Base(f) == "004_reservas_estados_pagos.sql" {
			sql = []byte(strings.Split(string(sql), "-- Job de pg_cron:")[0])
		}
		if _, err = pool.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	_, err = pool.Exec(ctx, `INSERT INTO usuarios(id,nombre,dpi,correo,contrasena,id_rol) VALUES
        (10,'Anfitrión','0000000000010','host@test','test',3),
        (11,'Otro anfitrión','0000000000011','other@test','test',3),
        (20,'Huésped','0000000000020','guest@test','test',2),
        (21,'Huésped 2','0000000000021','guest2@test','test',2),
        (22,'Huésped 3','0000000000022','guest3@test','test',2);
        INSERT INTO cabanas(id,nombre,direccion,precio,capacidad,id_anfitrion) VALUES
        (100,'Original','Dirección',100,4,10),(200,'Destino','Dirección',900,4,10),
        (300,'Ajena','Dirección',100,4,11),(400,'Pequeña','Dirección',100,2,10);`)
	if err != nil {
		t.Fatal(err)
	}
	return pool
}
func exec(t *testing.T, p *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := p.Exec(context.Background(), sql, args...); err != nil {
		t.Fatal(err)
	}
}
func number(t *testing.T, p *pgxpool.Pool, sql string, args ...any) float64 {
	t.Helper()
	var n float64
	if err := p.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
func request() maintenance.CreateRequest {
	return maintenance.CreateRequest{StartDate: "2099-01-02", EndDate: "2099-01-04", Reason: "Reparación"}
}
func reserve(t *testing.T, p *pgxpool.Pool, id, guest int, status string) {
	t.Helper()
	exec(t, p, `INSERT INTO reservaciones(id,id_cabana,id_usuario,fecha_inicio,fecha_fin,estado) VALUES($1,100,$2,'2099-01-01','2099-01-05',$3)`, id, guest, status)
}

func TestIntegrationGuestControlsResolution(t *testing.T) {
	p := fixture(t)
	ctx := context.Background()
	svc := maintenance.NewService(maintenance.NewRepository(p))
	reserve(t, p, 1000, 20, "confirmada")
	reserve(t, p, 1001, 21, "pendiente")
	exec(t, p, `INSERT INTO pagos(id_usuario,cantidad,tipo_pago,method,id_reservacion) VALUES(20,150,'anticipo','efectivo',1000)`)
	m, err := svc.Create(ctx, 100, 10, request())
	if err != nil {
		t.Fatal(err)
	}
	if number(t, p, `SELECT count(*) FROM reservaciones WHERE estado='por_mantenimiento'`) != 2 {
		t.Fatal("no suspendió las reservas")
	}
	if number(t, p, `SELECT count(*) FROM pagos WHERE tipo_pago='reembolso'`) != 0 {
		t.Fatal("devolvió sin consentimiento")
	}
	if number(t, p, `SELECT count(*) FROM notificaciones WHERE tipo='reserva_mantenimiento'`) != 2 {
		t.Fatal("no notificó a huéspedes")
	}
	if _, err := svc.Pending(ctx, 1000, 10); !errors.Is(err, maintenance.ErrForbidden) {
		t.Fatal("el anfitrión puede decidir", err)
	}
	d, err := svc.Pending(ctx, 1000, 20)
	if err != nil || d.TotalPrice != 400 || d.Nights != 4 || d.Refundable != 150 {
		t.Fatalf("solicitud: %+v %v", d, err)
	}
	if err := svc.Resolve(ctx, 1000, 21, maintenance.DecisionRequest{Action: "reembolsar"}); !errors.Is(err, maintenance.ErrForbidden) {
		t.Fatal("decidió otro huésped", err)
	}
	// Un cambio de tarifa no altera el precio que el huésped ya contrató.
	exec(t, p, `UPDATE cabanas SET precio=900 WHERE id=100`)
	if err := svc.Resolve(ctx, 1000, 20, maintenance.DecisionRequest{Action: "reprogramar", StartDate: "2099-02-01", EndDate: "2099-02-05"}); err != nil {
		t.Fatal(err)
	}
	if number(t, p, `SELECT count(*) FROM reservaciones WHERE id=1000 AND id_cabana=100 AND estado='confirmada' AND precio_total=400 AND fecha_inicio='2099-02-01'`) != 1 {
		t.Fatal("no conservó cabaña y total")
	}
	rr := reservations.NewRepository(p)
	cards, err := rr.GetReservationsByUserID(ctx, 20)
	if err != nil || cards[0].TotalPrice != 400 {
		t.Fatal(cards, err)
	}
	info, err := payments.NewRepository(p).GetReservationInfo(ctx, 1000)
	if err != nil || *info.TotalPrice != 400 {
		t.Fatal(info, err)
	}
	ledger, err := rr.GetReservationLedger(ctx, 1000, 100)
	if err != nil || *ledger.TotalPrice != 400 {
		t.Fatal(ledger, err)
	}
	if err := svc.Resolve(ctx, 1001, 21, maintenance.DecisionRequest{Action: "reembolsar"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Cancel(ctx, 100, m.ID, 10); err != nil {
		t.Fatal(err)
	}
	if number(t, p, `SELECT count(*) FROM reservaciones WHERE id=1001 AND estado='cancelada'`) != 1 {
		t.Fatal("cancelar bloque restauró una decisión ya tomada")
	}
}
func TestIntegrationRefundIsFullAndCannotRepeat(t *testing.T) {
	p := fixture(t)
	ctx := context.Background()
	svc := maintenance.NewService(maintenance.NewRepository(p))
	reserve(t, p, 1000, 20, "confirmada")
	exec(t, p, `INSERT INTO pagos(id_usuario,cantidad,tipo_pago,method,id_reservacion) VALUES
        (20,300,'anticipo','efectivo',1000),(10,50,'reembolso','transferencia',1000)`)
	if _, err := svc.Create(ctx, 100, 10, request()); err != nil {
		t.Fatal(err)
	}
	if err := svc.Resolve(ctx, 1000, 20, maintenance.DecisionRequest{Action: "reembolsar"}); err != nil {
		t.Fatal(err)
	}
	if number(t, p, `SELECT sum(cantidad) FROM pagos WHERE tipo_pago='reembolso' AND id_reservacion=1000`) != 300 {
		t.Fatal("reembolso no fue total neto")
	}
	if err := svc.Resolve(ctx, 1000, 20, maintenance.DecisionRequest{Action: "reembolsar"}); !errors.Is(err, maintenance.ErrNotFound) {
		t.Fatal("permitió decisión repetida", err)
	}
	if number(t, p, `SELECT sum(cantidad) FROM pagos WHERE tipo_pago='reembolso' AND id_reservacion=1000`) != 300 {
		t.Fatal("duplicó reembolso")
	}
}
func TestIntegrationUnavailableDatesKeepDecisionPending(t *testing.T) {
	p := fixture(t)
	ctx := context.Background()
	svc := maintenance.NewService(maintenance.NewRepository(p))
	reserve(t, p, 1000, 20, "confirmada")
	if _, err := svc.Create(ctx, 100, 10, request()); err != nil {
		t.Fatal(err)
	}
	exec(t, p, `INSERT INTO reservaciones(id_cabana,id_usuario,fecha_inicio,fecha_fin,estado) VALUES(100,21,'2099-02-01','2099-02-05','confirmada'),(200,20,'2099-03-01','2099-03-05','confirmada')`)
	for _, dates := range [][2]string{{"2099-01-01", "2099-01-05"}, {"2099-02-01", "2099-02-05"}, {"2099-03-01", "2099-03-05"}, {"2099-04-01", "2099-04-08"}} {
		if err := svc.Resolve(ctx, 1000, 20, maintenance.DecisionRequest{Action: "reprogramar", StartDate: dates[0], EndDate: dates[1]}); err == nil {
			t.Fatal("aceptó fechas inválidas", dates)
		}
	}
	if number(t, p, `SELECT count(*) FROM mantenimiento_reservaciones WHERE id_reservacion=1000 AND accion='pendiente'`) != 1 {
		t.Fatal("decisión dejó de estar pendiente")
	}
	if number(t, p, `SELECT precio_total FROM reservaciones WHERE id=1000`) != 400 {
		t.Fatal("cambió total durante fallo")
	}
}
func TestIntegrationCancelRestoresOnlyWhenAllBlocksGone(t *testing.T) {
	p := fixture(t)
	ctx := context.Background()
	svc := maintenance.NewService(maintenance.NewRepository(p))
	reserve(t, p, 1000, 20, "confirmada")
	reserve(t, p, 1001, 21, "pendiente")
	a, err := svc.Create(ctx, 100, 10, maintenance.CreateRequest{StartDate: "2099-01-02", EndDate: "2099-01-03", Reason: "A"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.Create(ctx, 100, 10, maintenance.CreateRequest{StartDate: "2099-01-03", EndDate: "2099-01-04", Reason: "B"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Cancel(ctx, 100, a.ID, 11); !errors.Is(err, maintenance.ErrForbidden) {
		t.Fatal("canceló otro anfitrión")
	}
	if _, err := svc.Cancel(ctx, 100, a.ID, 10); err != nil {
		t.Fatal(err)
	}
	if number(t, p, `SELECT count(*) FROM reservaciones WHERE id=1000 AND estado='por_mantenimiento'`) != 1 {
		t.Fatal("restauró con otro bloqueo")
	}
	if _, err := svc.Cancel(ctx, 100, b.ID, 10); err != nil {
		t.Fatal(err)
	}
	if number(t, p, `SELECT count(*) FROM reservaciones WHERE id=1000 AND estado='confirmada' AND precio_total=400`) != 1 {
		t.Fatal("no restituyó reserva")
	}
}
func TestIntegrationSeasonalTotalIsSharedAndStored(t *testing.T) {
	p := fixture(t)
	ctx := context.Background()
	exec(t, p, `INSERT INTO temporada(nombre,porcentaje_ajuste_sugerido,activo,fecha_inicio,fecha_fin) VALUES('Global',25,true,'2099-01-02','2099-01-05');
        INSERT INTO precios_temporada(id_cabana,fecha_inicio,fecha_fin,porcentaje_ajuste,precio_fijo) VALUES
        (100,'2099-01-03','2099-01-03',50,NULL),(100,'2099-01-04','2099-01-04',0,333.33),(100,'2099-01-05','2099-01-05',0,NULL);`)
	quote, err := seasons.NewService(seasons.NewRepository(p)).GetPriceQuote(ctx, seasons.PriceQuoteRequest{CabinID: 100, StartDate: "2099-01-01", EndDate: "2099-01-06"})
	if err != nil || quote.TotalPrice != 808.33 {
		t.Fatalf("cotización: %+v %v", quote, err)
	}
	rr := reservations.NewRepository(p)
	id, err := rr.CreateReservation(ctx, reservations.ReservationRequest{UserID: 20, CabinID: 100, StartDate: "2099-01-01", EndDate: "2099-01-06"})
	if err != nil {
		t.Fatal(err)
	}
	detail, err := rr.GetReservationByID(ctx, id)
	if err != nil || *detail.TotalPrice != 808.33 {
		t.Fatal(detail, err)
	}
	_, amount, err := reservations.NewService(rr, nil).PayAnticipo(ctx, id, 0, "efectivo")
	if err != nil || amount != 242.50 {
		t.Fatalf("anticipo %v: %v", amount, err)
	}
	exec(t, p, `UPDATE cabanas SET precio=500 WHERE id=100; UPDATE temporada SET porcentaje_ajuste_sugerido=80 WHERE nombre='Global'`)
	detail, err = rr.GetReservationByID(ctx, id)
	if err != nil || *detail.TotalPrice != 808.33 {
		t.Fatal("cambió precio ya contratado", detail, err)
	}
	balance, err := reservations.NewService(rr, nil).GetReservationBalance(ctx, detail)
	if err != nil || balance.BaseTotal != 808.33 || balance.PendingBalance != 565.83 {
		t.Fatalf("saldo: %+v %v", balance, err)
	}
}

func TestIntegrationConcurrentDecisionsOnlyOneRefund(t *testing.T) {
	p := fixture(t)
	ctx := context.Background()
	svc := maintenance.NewService(maintenance.NewRepository(p))
	reserve(t, p, 1000, 20, "confirmada")
	exec(t, p, `INSERT INTO pagos(id_usuario,cantidad,tipo_pago,method,id_reservacion) VALUES(20,150,'anticipo','efectivo',1000)`)
	if _, err := svc.Create(ctx, 100, 10, request()); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results <- svc.Resolve(ctx, 1000, 20, maintenance.DecisionRequest{Action: "reembolsar"})
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, maintenance.ErrNotFound) {
			t.Fatal(err)
		}
	}
	if success != 1 || number(t, p, `SELECT sum(cantidad) FROM pagos WHERE id_reservacion=1000 AND tipo_pago='reembolso'`) != 150 {
		t.Fatal("decisiones simultáneas duplicaron la devolución")
	}
}
func TestIntegrationSnapshotCannotBeChosenByClientAndEditsRequote(t *testing.T) {
	p := fixture(t)
	ctx := context.Background()
	exec(t, p, `INSERT INTO reservaciones(id,id_cabana,id_usuario,fecha_inicio,fecha_fin,precio_total) VALUES(1000,100,20,'2099-01-01','2099-01-05',1)`)
	if number(t, p, `SELECT precio_total FROM reservaciones WHERE id=1000`) != 400 {
		t.Fatal("aceptó total impuesto")
	}
	exec(t, p, `UPDATE cabanas SET precio=123.45 WHERE id=100`)
	if err := reservations.NewRepository(p).UpdateReservation(ctx, 1000, reservations.UpdateReservationRequest{CabinID: 100, StartDate: "2099-02-01", EndDate: "2099-02-05"}); err != nil {
		t.Fatal(err)
	}
	if number(t, p, `SELECT precio_total FROM reservaciones WHERE id=1000`) != 493.80 {
		t.Fatal("no recotizó edición normal")
	}
}

func TestIntegrationUpgradeBackfillsExistingReservations(t *testing.T) {
	p := fixture(t, func(p *pgxpool.Pool) {
		exec(t, p, `INSERT INTO cabanas(id,nombre,direccion,precio,capacidad,id_anfitrion) VALUES(500,'Histórica','Dirección',100,4,1);
            INSERT INTO temporada(nombre,porcentaje_ajuste_sugerido,activo,fecha_inicio,fecha_fin) VALUES('Prueba histórica',25,true,'2099-01-01','2099-01-04');
            INSERT INTO reservaciones(id,id_cabana,id_usuario,fecha_inicio,fecha_fin,precio_total_acordado,precio_noche_acordado) VALUES
            (500,500,1,'2099-01-01','2099-01-05',777,NULL),(501,500,1,'2099-01-01','2099-01-05',NULL,123.45),(502,500,1,'2099-01-01','2099-01-05',NULL,NULL);`)
	})
	for id, want := range map[int]float64{500: 777, 501: 493.8, 502: 500} {
		if got := number(t, p, `SELECT precio_total FROM reservaciones WHERE id=$1`, id); got != want {
			t.Fatalf("migración reserva %d: %v, esperaba %v", id, got, want)
		}
	}
}

func TestIntegrationAbonosAfterAnticipoThroughHTTP(t *testing.T) {
	p := fixture(t)
	ctx := context.Background()
	t.Setenv("JWT_SECRET", "payment-integration-test")
	exec(t, p, `UPDATE cabanas SET precio=100.01 WHERE id=100`)
	rr := reservations.NewRepository(p)
	id, err := rr.CreateReservation(ctx, reservations.ReservationRequest{UserID: 20, CabinID: 100, StartDate: "2099-02-01", EndDate: "2099-02-05"})
	if err != nil {
		t.Fatal(err)
	}
	if _, amount, err := reservations.NewService(rr, nil).PayAnticipo(ctx, id, 0, "tarjeta"); err != nil || amount != 120.01 {
		t.Fatal("anticipo", amount, err)
	}
	mux := http.NewServeMux()
	payments.NewController(payments.NewService(payments.NewRepository(p), nil)).RegisterRoutes(mux)
	handler := middleware.Auth(mux)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": 20, "role": 2}).SignedString([]byte("payment-integration-test"))
	if err != nil {
		t.Fatal(err)
	}
	for _, payment := range []struct {
		amount string
		status int
	}{{"140.01", 201}, {"140.02", 201}, {"0.01", 409}} {
		req := httptest.NewRequest("POST", "/cdv-api/payments/reservation/"+strconv.Itoa(id), strings.NewReader(`{"amount":`+payment.amount+`,"method":"tarjeta"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != payment.status {
			t.Fatalf("abono %s: %d %s", payment.amount, response.Code, response.Body)
		}
	}
	req := httptest.NewRequest("GET", "/cdv-api/payments/reservation/"+strconv.Itoa(id), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	var summary payments.ReservationPayments
	if err := json.Unmarshal(response.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || summary.Total != 400.04 || summary.Paid != 400.04 || summary.Balance != 0 || len(summary.Payments) != 3 {
		t.Fatalf("saldo final: %d %+v", response.Code, summary)
	}
}
