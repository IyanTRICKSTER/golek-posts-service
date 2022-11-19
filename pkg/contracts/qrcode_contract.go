package contracts

type QrCodeRepository interface {
	Generate(text string) (url string, err error)
}