package repositories

import "golek_posts_service/pkg/contracts"

type QRCode struct {
}

func (q QRCode) Generate(text string) (url string, err error) {
	return "https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=" + text, nil
}

func NewQRCodeRepository() contracts.QrCodeRepository {
	return &QRCode{}
}
