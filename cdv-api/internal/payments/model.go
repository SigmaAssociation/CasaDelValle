package payments

import (
	"errors"
	"time"
)

// Payment es un pago o abono registrado sobre una reservación. Los pagos son
// simulados: no se procesa dinero real, solo se deja constancia del monto, el
// método, la fecha y la reservación relacionada.
type Payment struct {
	ID            int       `json:"id"`
	ReservationID int       `json:"reservationId"`
	UserID        uint      `json:"userId"`
	Amount        float64   `json:"amount"`
	Method        string    `json:"method"`
	Date          time.Time `json:"date"`
}

// Métodos de pago. "anticipo" ya existe como registro histórico (el primer
// pago que confirma la reservación); los demás son los abonos que puede
// registrar el usuario.
const (
	MethodAnticipo = "anticipo"
	MethodCard     = "tarjeta"
	MethodTransfer = "transferencia"
	MethodCash     = "efectivo"
)

// payableMethods son los métodos que un usuario puede elegir al registrar un
// pago manualmente (el anticipo lo genera el sistema, no se registra a mano).
var payableMethods = map[string]bool{
	MethodCard:     true,
	MethodTransfer: true,
	MethodCash:     true,
}

// Códigos del medio de pago que se guardan en la columna `method` de la tabla
// `pagos` (diseño de la base de datos: tipo_pago distingue anticipo/pago y
// method el medio). La API expone valores amigables ("tarjeta",
// "transferencia", "efectivo"); la base guarda estos códigos.
const (
	dbMethodCard     = "CARD_SIMULATED"
	dbMethodTransfer = "TRANSFER_SIMULATED"
	dbMethodCash     = "CASH"
)

// dbMethodCode traduce el método de la API al código que se guarda en la base.
func dbMethodCode(method string) (string, bool) {
	switch method {
	case MethodCard:
		return dbMethodCard, true
	case MethodTransfer:
		return dbMethodTransfer, true
	case MethodCash:
		return dbMethodCash, true
	default:
		return "", false
	}
}

// apiMethod traduce un registro de la base al método que expone la API. Los
// pagos de tipo "anticipo" se muestran como tales; el resto según su código.
func apiMethod(tipoPago, methodCode string) string {
	if tipoPago == MethodAnticipo {
		return MethodAnticipo
	}
	switch methodCode {
	case dbMethodCard:
		return MethodCard
	case dbMethodTransfer:
		return MethodTransfer
	case dbMethodCash:
		return MethodCash
	default:
		return methodCode
	}
}

// Estados de la reservación donde se admite registrar un pago.
const (
	statusConfirmed  = "confirmada"
	statusInProgress = "en_curso"
)

// CreatePaymentRequest es el cuerpo para registrar un pago simulado.
type CreatePaymentRequest struct {
	Amount float64 `json:"amount"`
	Method string  `json:"method"`
}

// ReservationPaymentInfo reúne los datos de la reservación y su cabaña que se
// necesitan para validar un pago y calcular el saldo.
type ReservationPaymentInfo struct {
	TotalPrice    *float64
	ReservationID int
	UserID        uint
	HostID        uint
	CabinID       uint
	CabinName     string
	Status        string
	StartDate     string
	EndDate       string
	NightlyRate   float64
}

// ReservationPayments es el desglose del dinero de una reservación: total,
// pagado, saldo pendiente e historial de pagos.
type ReservationPayments struct {
	ReservationID int       `json:"reservationId"`
	Total         float64   `json:"total"`
	Paid          float64   `json:"paid"`
	Balance       float64   `json:"balance"`
	Payments      []Payment `json:"payments"`
}

// CreatePaymentResponse devuelve el pago recién registrado junto con el
// desglose actualizado de la reservación.
type CreatePaymentResponse struct {
	Message       string    `json:"message"`
	Payment       *Payment  `json:"payment"`
	ReservationID int       `json:"reservationId"`
	Total         float64   `json:"total"`
	Paid          float64   `json:"paid"`
	Balance       float64   `json:"balance"`
	Payments      []Payment `json:"payments"`
}

var (
	ErrReservationNotFound   = errors.New("reservation not found")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInvalidMethod         = errors.New("invalid payment method")
	ErrAmountExceedsBalance  = errors.New("amount exceeds pending balance")
	ErrReservationNotPayable = errors.New("reservation not payable")
)
