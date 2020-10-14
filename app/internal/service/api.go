package service

import (
	"avito/config"
	"avito/internal/storage"
)

type ServiceAPI interface {
	GetAdsService() AdsServiceAPI
	GetEmailService() EmailServiceAPI
	GetConfirmationService() ConfirmationServiceAPI
}

type serviceAPI struct {
	adsServiceAPI       AdsServiceAPI
	emailService        EmailServiceAPI
	confirmationService ConfirmationServiceAPI
}

func NewServiceAPI(api storage.StorageAPI, emailServerConfig *config.EmailServer, applicationConfig *config.ApplicationConfig) ServiceAPI {
	return &serviceAPI{
		adsServiceAPI:       NewAdsServiceAPI(api),
		emailService:        NewEmailService(emailServerConfig, applicationConfig),
		confirmationService: NewConfirmationServiceAPI(api),
	}
}

func (s *serviceAPI) GetAdsService() AdsServiceAPI {
	return s.adsServiceAPI
}

func (s *serviceAPI) GetEmailService() EmailServiceAPI {
	return s.emailService
}

func (s *serviceAPI) GetConfirmationService() ConfirmationServiceAPI {
	return s.confirmationService
}
