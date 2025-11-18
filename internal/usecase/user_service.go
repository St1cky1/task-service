package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/repository"
)

type UserService struct {
	userRepo   repository.IUserRepository
	avatarRepo repository.IAvatarRepository
}

func NewUserService(
	userRepo repository.IUserRepository,
	avatarRepo repository.IAvatarRepository,
) *UserService {
	return &UserService{
		userRepo:   userRepo,
		avatarRepo: avatarRepo,
	}
}

// CreateUser создает нового пользователя
func (s *UserService) CreateUser(ctx context.Context, req *entity.CreateUserRequest) (*entity.User, error) {
	user, err := s.userRepo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser получает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, userID int) (*entity.User, error) {
	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	return user, nil
}

// UpdateUser обновляет пользователя
func (s *UserService) UpdateUser(ctx context.Context, userID int, req *entity.UpdateUserRequest) (*entity.User, error) {
	// Проверяем что пользователь существует
	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}

	user, err = s.userRepo.Update(ctx, userID, updates)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, userID int) error {
	// Проверяем что пользователь существует
	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return entity.ErrUserNotFound
	}

	// Удаляем аватарку если существует
	avatar, err := s.avatarRepo.GetByUserId(ctx, userID)
	if err != nil {
		return err
	}
	if avatar != nil {
		os.Remove(avatar.FilePath)
		err := s.avatarRepo.DeleteByUserId(ctx, userID)
		if err != nil {
			return err
		}
	}

	// Удаляем пользователя
	err = s.userRepo.Delete(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

// ListUsers получает список пользователей
func (s *UserService) ListUsers(ctx context.Context) ([]entity.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// UploadAvatar загружает аватарку пользователя
func (s *UserService) UploadAvatar(ctx context.Context, userID int, data []byte, contentType string) (string, error) {
	// Проверяем что пользователь существует
	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", entity.ErrUserNotFound
	}

	// Проверяем размер файла (максимум 5MB)
	if len(data) > 5*1024*1024 {
		return "", fmt.Errorf("file size exceeds 5MB limit")
	}

	// Создаем директорию если её нет
	uploadDir := "var/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Генерируем имя файла
	fileName := fmt.Sprintf("avatar_%d_%d", userID, time.Now().UnixNano())
	filePath := filepath.Join(uploadDir, fileName)

	// Сохраняем файл
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", err
	}

	// Удаляем старую аватарку если существует
	oldAvatar, err := s.avatarRepo.GetByUserId(ctx, userID)
	if err == nil && oldAvatar != nil {
		os.Remove(oldAvatar.FilePath)
	}

	// Сохраняем информацию об аватарке в БД
	avatar := &entity.Avatar{
		UserID:      userID,
		FilePath:    filePath,
		FileSize:    len(data),
		ContentType: contentType,
	}

	_, err = s.avatarRepo.Save(ctx, avatar)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	// Обновляем поле avatar_url в пользователе
	updates := make(map[string]interface{})
	updates["avatar_url"] = filePath

	_, err = s.userRepo.Update(ctx, userID, updates)
	if err != nil {
		os.Remove(filePath)
		s.avatarRepo.DeleteByUserId(ctx, userID)
		return "", err
	}

	return filePath, nil
}

// DownloadAvatar скачивает аватарку пользователя
func (s *UserService) DownloadAvatar(ctx context.Context, userID int) ([]byte, string, error) {
	// Проверяем что пользователь существует
	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", entity.ErrUserNotFound
	}

	// Получаем информацию об аватарке
	avatar, err := s.avatarRepo.GetByUserId(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if avatar == nil {
		return nil, "", fmt.Errorf("avatar not found")
	}

	// Читаем файл
	data, err := os.ReadFile(avatar.FilePath)
	if err != nil {
		return nil, "", err
	}

	return data, avatar.ContentType, nil
}

// DownloadAvatarStream скачивает аватарку с использованием stream
func (s *UserService) DownloadAvatarStream(ctx context.Context, userID int, chunkSize int) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte)
	errChan := make(chan error, 1)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		// Проверяем что пользователь существует
		user, err := s.userRepo.GetById(ctx, userID)
		if err != nil {
			errChan <- err
			return
		}
		if user == nil {
			errChan <- entity.ErrUserNotFound
			return
		}

		// Получаем информацию об аватарке
		avatar, err := s.avatarRepo.GetByUserId(ctx, userID)
		if err != nil {
			errChan <- err
			return
		}
		if avatar == nil {
			errChan <- fmt.Errorf("avatar not found")
			return
		}

		// Открываем файл
		file, err := os.Open(avatar.FilePath)
		if err != nil {
			errChan <- err
			return
		}
		defer file.Close()

		// Читаем и отправляем чанками
		buffer := make([]byte, chunkSize)
		for {
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				errChan <- err
				return
			}

			if n > 0 {
				select {
				case dataChan <- buffer[:n]:
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				}
			}

			if err == io.EOF {
				break
			}
		}
	}()

	return dataChan, errChan
}

// HasAvatar проверяет, есть ли аватарка у пользователя
func (s *UserService) HasAvatar(ctx context.Context, userID int) bool {
	avatar, err := s.avatarRepo.GetByUserId(ctx, userID)
	return err == nil && avatar != nil
}
