package usecase

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	avatarDir          = "/Users/v.petrov/avatar"
	totalImages        = 5
	maxAutoUsers       = 30
	userCreateInterval = 30 * time.Second
)

// ContinuousUserGenerationWithAvatars генерирует пользователей каждые 30 секунд с аватарками
// Создает до maxAutoUsers пользователей
func ContinuousUserGenerationWithAvatars(ctx context.Context, userService *UserService) error {
	// Загружаем картинки с диска один раз
	log.Println("📷 Загрузка изображений аватарок с диска...")
	images, err := loadAvatarImages()
	if err != nil {
		return fmt.Errorf("❌ Ошибка загрузки аватарок: %w", err)
	}
	log.Printf("✅ Загружено %d изображений аватарок\n", len(images))

	log.Printf("\n👥 Начинаем непрерывную генерацию пользователей (максимум %d)...\n", maxAutoUsers)
	log.Printf("⏱️  Новый пользователь будет создаваться каждые %d сек\n\n", int(userCreateInterval.Seconds()))

	start := time.Now()
	userCount := 0
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	// Уникальный префикс на основе времени запуска для предотвращения дублирования email
	sessionID := time.Now().Unix()

	// Индекс для циклического выбора аватарки
	avatarIdx := 0

	ticker := time.NewTicker(userCreateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			duration := time.Since(start)
			separator := strings.Repeat("=", 60)
			fmt.Println("\n" + separator)
			fmt.Printf("📊 Статистика генерации пользователей:\n")
			fmt.Printf("   Создано пользователей:    %d\n", userCount)
			fmt.Printf("   Успешно с аватарками:     %d ✅\n", successCount)
			fmt.Printf("   Ошибок:                   %d ❌\n", errorCount)
			fmt.Printf("   Время работы:             %.2f сек\n", duration.Seconds())
			fmt.Println(separator)
			fmt.Println("🛑 Генерация пользователей остановлена")
			return nil

		case <-ticker.C:
			mu.Lock()
			if userCount >= maxAutoUsers {
				mu.Unlock()
				fmt.Printf("✅ Достигнут максимум пользователей (%d). Генерация завершена\n", maxAutoUsers)
				return nil
			}

			userCount++
			currentUserNum := userCount
			mu.Unlock()

			// Запускаем создание пользователя в горутине
			go func(userNum int, imgIdx int, sid int64) {
				// Генерируем пароль
				password := generateRandomPassword(12)

				// Генерируем email и имя с уникальным sessionID
				email := fmt.Sprintf("auto_user_%d_%d@task-service.local", sid, userNum)
				name := fmt.Sprintf("Auto User %d", userNum)

				// Создаем контекст с timeout
				ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()

				// Создаем пользователя с аватаркой
				user, err := userService.CreateUserWithAvatar(ctxWithTimeout, name, email, password, images[imgIdx])

				mu.Lock()
				if err != nil {
					log.Printf("❌ User %2d: Ошибка создания - %v\n", userNum, err)
					errorCount++
				} else {
					log.Printf("✅ User %2d: Создан успешно (ID=%d, Email=%s, Пароль=%s)\n", userNum, user.ID, email, password)
					successCount++
				}
				mu.Unlock()
			}(currentUserNum, avatarIdx, sessionID)

			// Переходим к следующей аватарке (циклически)
			avatarIdx = (avatarIdx + 1) % len(images)
		}
	}
}

// generateRandomPassword генерирует случайный пароль указанной длины
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	password := make([]byte, length)
	for i := range password {
		password[i] = charset[seed.Intn(len(charset))]
	}
	return string(password)
}

// loadAvatarImages загружает картинки из директории
func loadAvatarImages() ([][]byte, error) {
	var images [][]byte

	for i := 1; i <= totalImages; i++ {
		filename := fmt.Sprintf("%d.jpeg", i)
		filePath := filepath.Join(avatarDir, filename)

		// Читаем файл
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения файла %s: %w", filename, err)
		}

		images = append(images, data)
	}

	return images, nil
}

// uploadUserAvatar загружает аватарку для одного пользователя
func uploadUserAvatar(ctx context.Context, userService *UserService, userID int, imageData []byte) error {
	_, err := userService.UploadAvatar(ctx, userID, imageData, "image/jpeg")
	if err != nil {
		return fmt.Errorf("ошибка загрузки аватарки: %w", err)
	}
	return nil
}
