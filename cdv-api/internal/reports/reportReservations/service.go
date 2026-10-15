package reportreservations

import (
	"context"
	"errors"
)

const mensajeSinDatos = "No se encontraron reservaciones para los filtros seleccionados."

var ErrFiltroInvalido = errors.New("los filtros deben ser identificadores positivos")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GenerarReporte(ctx context.Context, filtro FiltroReporte) (*ReporteReservaciones, error)  {
	if (filtro.IDAnfitrion != nil && *filtro.IDAnfitrion <= 0) ||
		(filtro.IDCabana != nil && *filtro.IDCabana <= 0) {
		return nil, ErrFiltroInvalido
	}

	cabanas, err := s.repo.ObtenerReporte(ctx, filtro)
	if err != nil {
		return nil, err
	}

	reporte := &ReporteReservaciones{
		Cabanas:      cabanas,
		TotalCabanas: len(cabanas),
	}
	for _, c := range cabanas {
		reporte.TotalReservaciones += c.TotalReservaciones
	}

	if reporte.TotalReservaciones == 0 {
		reporte.Mensaje = mensajeSinDatos
	}

	return reporte, nil
}