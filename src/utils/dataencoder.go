package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// Соль для шифрования и дешифровки 32 символа
// Пример
// export CEPH_API_SALT=cfd8a6ba-1df9-4b96-932b-6362e230
// SALT будет встроена при компиляции через -ldflags
// берется из .env сборочной линии через контейнер
var SALT = ""

// Метод шифрования файла (AES-GCM)
func EncodeFile(data []byte) ([]byte, error) {
	key := []byte(SALT) // Ключ AES берётся напрямую из Salt без расширения/уменьшения

	block, e := aes.NewCipher(key) // Инициализация блока cipher
	if e != nil {
		return nil, e
	}

	aesGCM, e := cipher.NewGCM(block) // Создание инициалаизированного GCM инстанса (для AES-GCM режим шифрования).
	if e != nil {
		return nil, e
	}

	nonce := make([]byte, aesGCM.NonceSize()) // Генерация уникального nonce на случайное значение из rand.Reader для каждого файла

	if _, e = io.ReadFull(rand.Reader, nonce); e != nil { // Заполнение массива nonce рандомными байтами. Громоздкий подход
		return nil, e
	}

	//Encrypt data
	encodetext := aesGCM.Seal(nonce, nonce, data, nil) // Шифрование данных: возвращает nonce + зашифрованный текст

	return encodetext, nil
}

// Метод дешифрования файла (AES-GCM)
func DecodeFile(data []byte) ([]byte, error) {
	key := []byte(SALT)
	block, e := aes.NewCipher(key) // Инициализация блока cipher
	if e != nil {
		return nil, e
	}

	aesGCM, e := cipher.NewGCM(block) // Создание инициалаизированного GCM инстанса (для AES-GCM режим шифрования).
	if e != nil {
		return nil, e
	}

	nonceSize := aesGCM.NonceSize() // Определение размера nonce

	// Выделение из входящих данных массива nonce + зашифрованного текста (так как Seal возвращает и тот и другой)
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, e := aesGCM.Open(nil, nonce, ciphertext, nil) // Дешифрование: проверяет подлинность данных. Если не верно - вернёт ошибку.
	if e != nil {
		return nil, e
	}

	return plaintext, nil
}
