package driver

type GetPassworder interface {
	GetPassword(ip, user string) (string, error)
}
