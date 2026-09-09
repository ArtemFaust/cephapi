package uutils

// ChunkBy разбивает указатель на массив строк (*[]string) на части указанного размера.
func ChunkBy(items *[]string, chunkSize int) (chunks []*[]string) {
	// Если указатель nil или пустой — возвращаем пустые чанки.
	if items == nil || len(*items) <= 0 {
		return chunks
	}

	slice := *items // Копируем слайс из указателя для модификации в цикле

	for chunkSize < len(slice) {
		chunk := slice[:chunkSize]      // Вырезаем чанк: [start:end), где start=0 (по умолчанию) и end=chunkSize
		chunks = append(chunks, &chunk) // Добавляем ссылку на вырезанный подсписок в результирующий список

		slice = slice[chunkSize:] // Сдвигаем слайс: удаляем обработанную часть (начинаем со chunkSize-го индекса)
	}

	if len(slice) > 0 {
		chunks = append(chunks, &slice) // Добавляем последний оставшийся остаток как отдельный чанк
	}

	return chunks
}
