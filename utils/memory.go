package utils

import "sync"

var (
	codeStore = make(map[string]string)
	mu        = sync.Mutex{}
)

//Depict a db, a dummy storage

func SaveVerificationCode(email, code string) {
	mu.Lock()
	defer mu.Unlock()
	codeStore[email] = code
}

func VerifyCode(email, code string) bool {
	mu.Lock()
	defer mu.Unlock()
	if stored, ok := codeStore[email]; ok && stored == code {
		delete(codeStore, email) // one-time use
		return true
	}
	return false
}
