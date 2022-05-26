package utils

import (
	"fmt"
)

func ErrorHandling(e error) {
	if e != nil {
		fmt.Println(fmt.Errorf("Erro encontrado: %s, \n\nerror: %w", e.Error(), e))
	}
}
