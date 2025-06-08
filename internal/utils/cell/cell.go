package cell

import (
	"fmt"
	"strconv"
	"strings"
)

//columnNumber преобразует буквенное обозначение в номер столбца

func ColumnNumber(s string) int {
	result := 0
	for _, r := range s {
		digit := int(r - 'A' + 1)
		result = result*26 + digit
	}
	return result
}
func ColumnChar(n int) string {
	result := ""
	for n > 0 {
		n-- // Важно!
		result = string(rune('A'+(n%26))) + result
		n /= 26
	}
	return result
}
func ColumnsCountToAddresses(count int) []string {
	addresses := make([]string, count)
	for i := range count {
		addresses[i] = ColumnChar(i + 1)
	}
	return addresses
}

// parseCell парсит ячейку в номер столбца и строку
func ParseCell(cell string) (string, int, error) {
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
