package service

import (
	"avito/config"
	"fmt"
	"github.com/gofrs/uuid"
	"net/smtp"
)

type EmailServiceAPI interface {
	SendEmailChangePrice(receiver string, name string, price int) error
	SendEmailConfirmation(receiver string, name string, confirmationID uuid.UUID) error
	SendEmailConfirmed(receiver string, name string) error
	SendEmailClosedAd(receiver string, name string, link string) error
	SendEmailUnsubscription(receiver string, link string) error
}

type EmailSender struct {
	config             *config.EmailServer
	sendMailConfig     string
	confirmationConfig string
}

func NewEmailService(emailServerConfig *config.EmailServer, applicationConfig *config.ApplicationConfig) EmailServiceAPI {
	return &EmailSender{
		config:             emailServerConfig,
		sendMailConfig:     fmt.Sprintf("%s:%d", emailServerConfig.Server, emailServerConfig.Port),
		confirmationConfig: fmt.Sprintf("%s:%d", applicationConfig.Domain, applicationConfig.HTTPPort),
	}
}

func (e *EmailSender) SendEmailConfirmation(receiver string, name string, confirmationID uuid.UUID) error {
	return e.sendMail(receiver, e.getEmailTitleConfirmation(name), e.getEmailBodyConfirmation(name, confirmationID))
}

func (e *EmailSender) SendEmailConfirmed(receiver string, name string) error {
	return e.sendMail(receiver, e.getEmailTitleConfirmation(name), e.getEmailBodyConfirmedSubscription(name))
}

func (e *EmailSender) SendEmailChangePrice(receiver string, name string, price int) error {
	return e.sendMail(receiver, e.getEmailTitleChangePrice(name), e.getEmailBodyChangePrice(name, price))
}

func (e *EmailSender) SendEmailClosedAd(receiver string, name string, link string) error {
	return e.sendMail(receiver, e.getEmailTitleAdRemoved(name), e.getEmailBodyAdRemoved(name, link))
}

func (e *EmailSender) SendEmailUnsubscription(receiver string, link string) error {
	return e.sendMail(receiver, e.getEmailTitleUnsubscription(), e.getEmailBodyUnsubscription(link))
}

func (e *EmailSender) sendMail(receiver string, subject string, msgBody string) error {
	auth := smtp.PlainAuth(
		"",
		e.config.Name,
		e.config.Password,
		e.config.Server,
	)

	msg := "From: " + e.config.Name + "\r\n" +
		"To: " + receiver + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		msgBody

	err := smtp.SendMail(e.sendMailConfig, auth, e.config.Name, []string{receiver}, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

func (e *EmailSender) getEmailBodyChangePrice(name string, price int) string {
	return fmt.Sprintf("Изменена цена на отслеживаемое Вами объявление \"%s\". Новая цена: %d руб.", name, price)
}

func (e *EmailSender) getEmailBodyConfirmation(name string, confirmationID uuid.UUID) string {
	return fmt.Sprintf("Для подтверждения подписки на изменение цены объявления \"%s\" перейдите по ссылке %s/confirm?id=%s", name, e.confirmationConfig, uuid.UUID.String(confirmationID))
}

func (e *EmailSender) getEmailBodyConfirmedSubscription(name string) string {
	return fmt.Sprintf("Подписка на изменение цены объявления \"%s\" подтверждена!", name)
}

func (e *EmailSender) getEmailBodyAdRemoved(name string, link string) string {
	return fmt.Sprintf("Объявление \"%s\"(%s) было снято с публикации.", name, link)
}

func (e *EmailSender) getEmailBodyUnsubscription(link string) string {
	return fmt.Sprintf("Подписка на отслеживание цены объявления %s была отменена.", link)
}

func (e *EmailSender) getEmailTitleConfirmation(name string) string {
	return fmt.Sprintf("Подтверждение подписки на объявление \"%s\"", name)
}

func (e *EmailSender) getEmailTitleChangePrice(name string) string {
	return fmt.Sprintf("Изменение цены на товар \"%s\"", name)
}

func (e *EmailSender) getEmailTitleAdRemoved(name string) string {
	return fmt.Sprintf("Закрытие объявления \"%s\"", name)
}

func (e *EmailSender) getEmailTitleUnsubscription() string {
	return fmt.Sprint("Отписка от отслеживания объявления")
}
