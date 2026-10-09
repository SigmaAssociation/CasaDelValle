package users

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"cdv-api/internal/middleware"
	"cdv-api/internal/notifications"
)

// Errores de registro con mensajes amigables para el usuario. Se declaran
// como centinelas para que el controlador pueda decidir el código HTTP sin
// comparar cadenas de texto.
var (
	ErrDuplicateUser  = errors.New("El correo o DPI ya está registrado")
	ErrDuplicateEmail = errors.New("El correo ya está registrado")
	ErrDuplicateDPI   = errors.New("El DPI ya está registrado")
	ErrInvalidRole    = errors.New("El tipo de cuenta seleccionado no está disponible. Verifica la opción e intenta de nuevo")
	ErrInvalidData    = errors.New("Alguno de los datos no cumple con el formato permitido")
)

type Service struct {
	repository UserRepository
	notifier   Notifier
}

type Notifier interface {
	Notify(ctx context.Context, input notifications.NotificationInput)
	NotifyAdmins(ctx context.Context, input notifications.NotificationInput)
}

// UserRepository abstrae el acceso a datos de usuarios para poder sustituirlo
// en las pruebas unitarias.
type UserRepository interface {
	CreateUser(ctx context.Context, req RegisterRequest, hashPassword string) (int, error)
	UpdateUser(ctx context.Context, userUpdate UserUpdate) (int, error)
	UpdateRole(ctx context.Context, userID int, role uint) (int, error)
	GetAll(ctx context.Context) ([]User, error)
	GetUserAuthInfo(ctx context.Context, email string) (UserAuth, error)
	GetUserByID(ctx context.Context, id int) (User, error)
}

func NewService(repository UserRepository, notifier Notifier) *Service {
	return &Service{
		repository: repository,
		notifier:   notifier,
	}
}

func (s *Service) notify(ctx context.Context, userID int, tipo, mensaje, entidadTipo string, entidadID int) {
	if s.notifier == nil || userID <= 0 {
		return
	}
	s.notifier.Notify(ctx, notifications.NotificationInput{
		UserID:      uint(userID),
		Tipo:        tipo,
		Mensaje:     mensaje,
		EntidadTipo: entidadTipo,
		EntidadID:   entidadID,
	})
}

// notifyAdmins replica un evento del sistema en el buzón de los administradores.
func (s *Service) notifyAdmins(ctx context.Context, tipo, mensaje string, entidadID int) {
	if s.notifier == nil {
		return
	}
	s.notifier.NotifyAdmins(ctx, notifications.NotificationInput{
		Tipo:        tipo,
		Mensaje:     mensaje,
		EntidadTipo: notifications.EntidadUsuario,
		EntidadID:   entidadID,
	})
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
	if req.IDRole == 0 {
		req.IDRole = middleware.RoleGuest
	}
	if req.IDRole != middleware.RoleGuest && req.IDRole != middleware.RoleHost {
		return 0, errors.New("Rol inválido: debe ser huésped o anfitrión")
	}
	req.Address = strings.TrimSpace(req.Address)
	if req.Address != "" {
		if len(req.Address) < 5 || len(req.Address) > 255 {
			return 0, errors.New("Dirección inválida: debe tener entre 5 y 255 caracteres")
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$`, req.Address); !matched {
			return 0, errors.New("Dirección inválida: contiene caracteres no permitidos")
		}
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost) // Hasheo de contraseña
	if err != nil {
		return 0, errors.New("Error al procesar la contraseña")
	}

	userID, err := s.repository.CreateUser(ctx, req, string(hashBytes))
	if err != nil {
		return 0, err
	}

	// Notificación de bienvenida al buzón personal del nuevo usuario.
	s.notify(ctx, userID, notifications.TipoBienvenida,
		fmt.Sprintf("¡Bienvenido a Casa Del Valle, %s! Gracias por registrarte.", req.Name),
		notifications.EntidadUsuario, userID)

	s.notifyAdmins(ctx, notifications.TipoBienvenida,
		fmt.Sprintf("Nuevo usuario registrado: %s (#%d).", req.Name, userID), userID)

	return userID, nil
}

func (s *Service) UpdateUser(ctx context.Context, req UserUpdate) (int, error) {

	if req.ID <= 0 {
		return 0, errors.New("ID de usuario inválido")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 150 {
		return 0, errors.New("Nombre inválido: debe tener entre 1 y 150 caracteres")
	}
	if matched, _ := regexp.MatchString(`^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$`, req.Name); !matched {
		return 0, errors.New("Nombre inválido: solo se permiten letras y espacios")
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if len(req.Email) > 150 || !emailRegex.MatchString(req.Email) {
		return 0, errors.New("Correo inválido: debe ser un correo electrónico válido y tener un máximo de 150 caracteres")
	}

	req.Phone = strings.TrimSpace(req.Phone)
	if matched, _ := regexp.MatchString(`^[123456789]\d{7}$`, req.Phone); !matched {
		return 0, errors.New("Teléfono inválido: debe contener 8 dígitos y no iniciar con 0")
	}

	req.Address = strings.TrimSpace(req.Address)
	if req.Address != "" {
		if len(req.Address) < 5 || len(req.Address) > 255 {
			return 0, errors.New("Dirección inválida: debe tener entre 5 y 255 caracteres")
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.\-#]+$`, req.Address); !matched {
			return 0, errors.New("Dirección inválida: contiene caracteres no permitidos")
		}
	}

	rowsAffected, err := s.repository.UpdateUser(ctx, req)
	if err != nil {
		return 0, err
	}

	if rowsAffected == 0 {
		return 0, errors.New("Usuario no encontrado")
	}

	return rowsAffected, nil
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

func (s *Service) GetUserByID(ctx context.Context, id int) (User, error) { // Pendiente de verificar si el usuario tiene permisos para acceder a la información del usuario con el ID proporcionado.
	// Esto podría implicar verificar el rol del usuario autenticado y compararlo con el ID del usuario solicitado.
	// Si el usuario no tiene permisos, se debería retornar un error de autorización.
	user, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, errors.New("Usuario no encontrado")
		}
		return User{}, err
	}
	return user, nil
}

// UpgradeToHost convierte un huésped en anfitrión. Es de una sola vía:
// no se permite volver a huésped ni cambiar el rol de un administrador.
func (s *Service) UpgradeToHost(ctx context.Context, id int) (UserAuth, error) {
	user, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserAuth{}, errors.New("Usuario no encontrado")
		}
		return UserAuth{}, err
	}

	if user.IDRole == middleware.RoleHost {
		return UserAuth{}, errors.New("El usuario ya es anfitrión")
	}
	if user.IDRole != middleware.RoleGuest {
		return UserAuth{}, errors.New("El cambio de rol no está permitido para esta cuenta")
	}

	rowsAffected, err := s.repository.UpdateRole(ctx, id, middleware.RoleHost)
	if err != nil {
		return UserAuth{}, err
	}
	if rowsAffected == 0 {
		return UserAuth{}, errors.New("Usuario no encontrado")
	}

	s.notify(ctx, id, notifications.TipoRolActualizado,
		"¡Felicidades! Tu cuenta ahora es Anfitrión. Ya puedes publicar tus cabañas.",
		notifications.EntidadUsuario, id)

	s.notifyAdmins(ctx, notifications.TipoRolActualizado,
		fmt.Sprintf("El usuario %s (#%d) ahora es Anfitrión.", user.Name, id), id)

	return UserAuth{
		ID:     user.ID,
		Email:  user.Email,
		IDRole: middleware.RoleHost,
		Name:   user.Name,
	}, nil
}
