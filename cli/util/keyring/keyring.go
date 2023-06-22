package keyring

import kr "github.com/zalando/go-keyring"

func Open(serviceName string) (Keyring, error) {
	return Keyring{
		ServiceName:      serviceName,
		ErrNotFound:      kr.ErrNotFound,
		ErrSetDataTooBig: kr.ErrSetDataTooBig,
	}, nil
}

type Keyring struct {
	ServiceName      string
	ErrNotFound      error
	ErrSetDataTooBig error
}

func (k Keyring) Get(key string) (string, error) {
	secret, err := kr.Get(k.ServiceName, key)
	if err != nil {
		return "", err
	}
	return secret, nil
}

func (k Keyring) Set(key string, value string) error {
	return kr.Set(k.ServiceName, key, value)
}

func (k Keyring) Delete(key string) error {
	return kr.Delete(k.ServiceName, key)
}
