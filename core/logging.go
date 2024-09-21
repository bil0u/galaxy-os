package core

import "fmt"

func LogInfo(message string) {
	fmt.Println("[INFO]", message)
}

func LogWarning(message string) {
	fmt.Println("[WARN]", message)
}

func LogError(message string) {
	fmt.Println("[ERR!]", message)
}
