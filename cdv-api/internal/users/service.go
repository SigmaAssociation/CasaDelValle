package users

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func (s *Service) AuthenticateUser(ctx context.Context, loginRequest UserLoginRequest) (string, error) {
	user, err := s.repository.GetUserAuthInfo(ctx, loginRequest.Email)
	if err != nil {
		return "", errors.New("Usuario no encontrado")
	}

	if (CheckPassword(loginRequest.Password, user.Password)) == false {
		return "", errors.New("Contraseña incorrecta")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := GenerateToken(user, jwtSecret)
	if err != nil {
		return "", errors.New("Error al generar el token")
	}

	return token, nil
}

func CheckPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	return err == nil
}

func GenerateToken(
	loginRequest UserAuth,
	secret string,
) (string, error) {

	claims := jwt.MapClaims{
		"sub":   loginRequest.ID,
		"email": loginRequest.Email,
		"role":  loginRequest.IDRole,
		"name":  loginRequest.Name,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(secret))
}

func (s *Service) GetUserByID(ctx context.Context, id int) (User, error) {
	// Pendiente de verificar si el usuario tiene permisos para acceder a la información del usuario con el ID proporcionado.
	// Esto podría implicar verificar el rol del usuario autenticado y compararlo con el ID del usuario solicitado.
	// Si el usuario no tiene permisos, se debería retornar un error de autorización.
	return s.repository.GetUserByID(ctx, id)
}
