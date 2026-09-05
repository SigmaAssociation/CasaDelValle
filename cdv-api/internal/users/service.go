package users

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetUsers(ctx context.Context) ([]User, error) {
	return s.repository.GetAll(ctx)
}

func (s *Service) RegisterUser(ctx context.Context, req RegisterRequest) (int, error) {
	if strings.TrimSpace(req.Name) == "" || len(req.Name) > 150 { // Validacion de nombre
		return 0, errors.New("Nombre inválido: debe tener entre 1 y 150 caracteres")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$`, req.Name); !matched {
		return 0, errors.New("Nombre inválido: solo se permiten letras y espacios")
	}
	if matched, _ := regexp.MatchString(`^\d{13}$`, req.DPI); !matched { // Validacion de DPI
		return 0, errors.New("DPI inválido: debe contener exactamente 13 dígitos")
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`) // Validacion de correo
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if len(req.Email) > 150 || !emailRegex.MatchString(req.Email) {
		return 0, errors.New("Correo inválido: debe ser un correo electrónico válido y tener un máximo de 150 caracteres")
	}
	if len(req.Password) < 8 { // Validacion de contraseña
		return 0, errors.New("Contraseña inválida: debe tener al menos 8 caracteres")
	}
	if hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(req.Password); !hasUpper {
		return 0, errors.New("Contraseña inválida: debe contener al menos una letra mayúscula")
	}
	if hasLower := regexp.MustCompile(`[a-z]`).MatchString(req.Password); !hasLower {
		return 0, errors.New("Contraseña inválida: debe contener al menos una letra minúscula")
	}
	if hasDigit := regexp.MustCompile(`\d`).MatchString(req.Password); !hasDigit {
		return 0, errors.New("Contraseña inválida: debe contener al menos un dígito")
	}
	if hasSpecial := regexp.MustCompile(`[!@#~$%^&*()+|_.,<>?/\\-]`).MatchString(req.Password); !hasSpecial {
		return 0, errors.New("Contraseña inválida: debe contener al menos un carácter especial")
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost) // Hasheo de contraseña
	if err != nil {
		return 0, errors.New("Error al procesar la contraseña")
	}
	return s.repository.CreateUser(ctx, req, string(hashBytes))
}
