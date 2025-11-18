package usecase

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	avatarDir   = "/Users/v.petrov/avatar"
	totalUsers  = 18
	totalImages = 5
)

// UploadAllAvatars загружает аватарки для всех пользователей
// Если у пользователя уже есть аватарка, её пропускает
func UploadAllAvatars(ctx context.Context, userService *UserService) error {
	// Загружаем картинки с диска
	log.Println(" Загрузка аватарок с диска...")
	images, err := loadAvatarImages()
	if err != nil {
		return fmt.Errorf("❌ Ошибка загрузки аватарок: %w", err)
	}
	log.Printf("✅ Загружено %d аватарок\n", len(images))

	log.Printf("\n Начинаем загрузку аватарок для %d пользователей...\n", totalUsers)

	start := time.Now()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 3) // Ограничиваем одновременные загрузки до 3

	successCount := 0
	skippedCount := 0
	errorCount := 0
	var mu sync.Mutex

	for userID := 1; userID <= totalUsers; userID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Проверяем, есть ли уже аватарка
			if userService.HasAvatar(ctx, id) {
				mu.Lock()
				log.Printf("⏭️  User %2d: Уже имеет аватарку, пропускаем\n", id)
				skippedCount++
				mu.Unlock()
				return
			}

			// Выбираем картинку циклически
			imageIdx := (id - 1) % len(images)
			imageData := images[imageIdx]

			// Загружаем аватарку
			ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			err := uploadUserAvatar(ctxWithTimeout, userService, id, imageData)

			mu.Lock()
			if err != nil {
				log.Printf("❌ User %2d: Ошибка загрузки - %v\n", id, err)
				errorCount++
			} else {
				log.Printf("✅ User %2d: Аватарка загружена успешно\n", id)
				successCount++
			}
			mu.Unlock()
		}(userID)
	}

	wg.Wait()
	duration := time.Since(start)

	// Выводим результаты
	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Printf("📊 Статистика загрузки аватарок:\n")
	fmt.Printf("   Всего пользователей: %d\n", totalUsers)
	fmt.Printf("   Успешно загружено:   %d ✅\n", successCount)
	fmt.Printf("   Пропущено:           %d ⏭️\n", skippedCount)
	fmt.Printf("   Ошибок:              %d ❌\n", errorCount)
	fmt.Printf("   Время:               %.2f сек\n", duration.Seconds())
	fmt.Println(separator)

	if errorCount > 0 {
		return fmt.Errorf("некоторые аватарки не удалось загрузить")
	}

	return nil
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
