package enc

import "sync"

var (
	Key = []byte("suckdaNaanddf394des239")
	mu  = &sync.Mutex{}
)

// 不足16位补0, 足的直接copy
func SetKey(s []byte) {
	mu.Lock()
	defer mu.Unlock()
	n := len(s)
	if n < 16 {
		t := 16 - n
		for t > 0 {
			s = append(s, '0')
			t--
		}
	} else {
		s = s[:16]
	}
	copy(Key, s)
}

// 获取Key
func GetKey() []byte {
	mu.Lock()
	defer mu.Unlock()
	return Key[:16]
}
