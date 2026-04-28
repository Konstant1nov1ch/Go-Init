package file

import (
	"log"
	"os"
	"runtime"
)

// CloseFile закрывает файл и логирует ошибку при неудаче.
func CloseFile(file *os.File) {
	if err := file.Close(); err != nil {
		log.Fatalf("Disk fatal error: %v", err)
	}
}

// OpenFile открывает файл, проверяя его существование.
func OpenFile(path string) (*os.File, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(file, func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Printf("Warning: failed to close file %s: %v", path, err)
		}
	})

	return file, nil
}
