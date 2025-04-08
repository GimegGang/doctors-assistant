package main

import (
	"fmt"
	"kode/internal/reception"
	"kode/internal/storage"
)

func main() {
	med := storage.Medicine{TakingDuration: 5}
	fmt.Println(reception.GetReceptionIntake(&med))
}
