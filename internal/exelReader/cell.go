package exelreader

import (
	"fmt"
	"strconv"
	"strings"
)

//columnNumber преобразует буквенное обозначение в номер столбца

func columnNumber(s string) int {
	result := 0
	for _, r := range s {
		digit := int(r - 'A' + 1)
		result = result*26 + digit
	}
	return result
}

// parseCell парсит ячейку в номер столбца и строку
func parseCell(cell string) (string, int, error) {
	// Разделяем на буквенную и числовую части
	letters := ""
	number := ""
	isNumber := false
	cell = strings.ToUpper(cell)
	for _, char := range cell {
		if char >= '0' && char <= '9' {
			isNumber = true
			number += string(char)
		} else {
			if isNumber {
				return "", 0, fmt.Errorf("некорректный формат ячейки")
			}
			if !(char >= 'A' && char <= 'Z') {
				return "", 0, fmt.Errorf("некорректный формат ячейки")
			}
			letters += string(char)
		}
	}
	if letters == "" || number == "" {
		return "", 0, fmt.Errorf("некорректный формат ячейки")
	}

	// Преобразуем строку
	row, err := strconv.Atoi(number)
	if err != nil {
		return "", 0, err
	}
	if row <= 0 {
		return "", 0, fmt.Errorf("некорректный номер строки")
	}
	return letters, row, nil
}
