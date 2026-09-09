package rgwapi

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/ceph/go-ceph/rgw/admin"
)

// Метод создания нового admin подключения к RGW API
func MakeNewConnection(url string, accessKey string, secretKey string) (*admin.API, error) {

	// Создание http клиента
	// Параметры TLS
	tlc := tls.Config{
		InsecureSkipVerify: true, // Отключаем проверку сертификата сервера
	}
	// параметры транспортного уровня
	transport := http.Transport{
		TLSClientConfig:       &tlc,
		IdleConnTimeout:       300 * time.Second, // Таймаут ожидания keep-alive простой соединения (после которого соединение будет считаться разорванным)
		ResponseHeaderTimeout: 300 * time.Second, // Таймаут чтения тела ответа
		TLSHandshakeTimeout:   300 * time.Second, // Таймаут установления TLS соединения
		MaxIdleConns:          0,
	}
	// параметры уровня протокола
	client := http.Client{
		Transport: &transport,
		Timeout:   480 * time.Second,
		// Проверка на кол-во редиректов
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse // Клиент не будет следовать за дальнейшими редиректами
			}
			return nil
		},
	}

	co, e := admin.New(url, accessKey, secretKey, &client)
	if e != nil {
		return nil, e
	}

	return co, nil
}
