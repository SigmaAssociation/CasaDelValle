package images

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

const (
	maxFileSize = 5 << 20 // 5 MB
	uploadBase  = "./uploads/cabins"
)

var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// CreateImage valida y guarda el archivo en disco y su ruta en la BD.
func (s *Service) CreateImage(ctx context.Context, cabinID, userID int, file multipart.File, header *multipart.FileHeader) (int, string, error) {
	if cabinID <= 0 {
		return 0, "", errors.New("ID de cabaña inválido")
	}

	// Validar tamaño
	if header.Size > maxFileSize {
		return 0, "", errors.New("La imagen excede el tamaño máximo de 5 MB")
	}

	// Validar MIME type real (leer primeros bytes)
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	file.Seek(0, io.SeekStart)
	mimeType := http.DetectContentType(buf[:n])
	ext, ok := allowedMimeTypes[mimeType]
	if !ok {
		return 0, "", errors.New("Formato de imagen no válido: solo JPEG, PNG o WebP")
	}

	// Crear directorio si no existe
	dir := filepath.Join(uploadBase, fmt.Sprintf("%d", cabinID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, "", errors.New("Error al crear directorio de subida")
	}

	// Nombre único para evitar colisiones (extensión derivada del tipo real)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(dir, filename)

	// Guardar en disco
	dst, err := os.Create(filePath)
	if err != nil {
		return 0, "", errors.New("Error al guardar la imagen")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return 0, "", errors.New("Error al escribir la imagen")
	}

	// Ruta relativa para guardar en BD (la que servirá el frontend)
	relPath := filepath.ToSlash(filepath.Join("uploads/cabins", fmt.Sprintf("%d", cabinID), filename))

	// Insertar en BD
	userIDPtr := &userID
	cabinIDPtr := &cabinID
	img := Image{
		Path:    relPath,
		UserID:  userIDPtr,
		CabinID: cabinIDPtr,
	}
	id, err := s.repository.CreateImage(ctx, img)
	if err != nil {
		// limpiar archivo si falla BD
		os.Remove(filePath)
		return 0, "", err
	}

	return id, relPath, nil
}

// GetImagesByCabin retorna todas las imágenes de una cabaña.
func (s *Service) GetImagesByCabin(ctx context.Context, cabinID int) (ImagesListResponse, error) {
	if cabinID <= 0 {
		return ImagesListResponse{}, errors.New("ID de cabaña inválido")
	}

	images, err := s.repository.GetByCabinID(ctx, cabinID)
	if err != nil {
		return ImagesListResponse{}, err
	}

	return ToResponseList(images), nil
}

// GetMainImageByCabin retorna la imagen más reciente de una cabaña.
func (s *Service) GetMainImageByCabin(ctx context.Context, cabinID int) (Image, error) {
	if cabinID <= 0 {
		return Image{}, errors.New("ID de cabaña inválido")
	}

	image, err := s.repository.GetMainByCabinID(ctx, cabinID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Image{}, errors.New("La cabaña no tiene imágenes registradas")
		}

		return Image{}, err
	}

	return image, nil
}

// GetImageByID retorna una imagen por su identificador.
func (s *Service) GetImageByID(ctx context.Context, id int) (Image, error) {
	if id <= 0 {
		return Image{}, errors.New("ID de imagen inválido")
	}

	image, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Image{}, errors.New("Imagen no encontrada")
		}

		return Image{}, err
	}

	return image, nil
}

// DeleteImage elimina el registro y también el archivo físico
// para no dejar archivos huérfanos.
func (s *Service) DeleteImage(ctx context.Context, img Image) (int, error) {
	if img.ID <= 0 {
		return 0, errors.New("ID de imagen inválido")
	}

	rows, err := s.repository.DeleteImage(ctx, img.ID)
	if err != nil {
		return 0, err
	}

	if rows == 0 {
		return 0, errors.New("Imagen no encontrada")
	}

	if img.Path != "" {
		os.Remove(img.Path)
	}

	return rows, nil
}
