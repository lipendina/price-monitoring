package service

import (
	"avito/internal"
	"avito/internal/utils"
	"context"
	"github.com/gofrs/uuid"
	"github.com/jasonlvhit/gocron"
	"golang.org/x/xerrors"
	"log"
	"os"
	"sync"
)

const limitAds = 1000

type MonitoringServiceAPI interface {
	Run()
}

type monitoring struct {
	service ServiceAPI
	log     *log.Logger
	isRun   bool
	wg      sync.WaitGroup
}

func NewMonitoring(api ServiceAPI) MonitoringServiceAPI {
	return &monitoring{
		service: api,
		log:     log.New(os.Stdout, "MONITORING: ", log.LstdFlags),
		isRun:   false,
	}
}

func (m *monitoring) executeJob() error {
	ctx := context.Background()
	if m.isRun {
		return xerrors.Errorf("Error, job is already run\n")
	}
	m.isRun = true

	ads, err := m.service.GetAdsService().GetAds(ctx, limitAds)
	if err != nil {
		return err
	}

	channelUpdated := make(chan *internal.Ad, limitAds)
	channelRemoved := make(chan *internal.Ad, limitAds)
	channelUpdateLastCheck := make(chan uuid.UUID, limitAds)
	for _, ad := range ads {
		m.wg.Add(1)
		go func(innerAd *internal.Ad) {
			defer m.wg.Done()
			parsedAd, err := utils.ParseAd(innerAd.Link)
			if err != nil {
				m.log.Printf("Error while parse ad, reason: %v\n", err)
				return
			}

			channelUpdateLastCheck <- innerAd.ID

			if parsedAd.Removed {
				channelRemoved <- innerAd
				return
			}

			if parsedAd.Price != innerAd.Price {
				innerAd.Price = parsedAd.Price
				channelUpdated <- innerAd
			}
		}(ad)
	}

	go func() {
		for {
			ad, ok := <-channelRemoved
			if !ok {
				break
			}
			ctx := context.Background()

			err := m.service.GetAdsService().UpdateAdIsRemoved(ctx, ad.ID)
			if err != nil {
				m.log.Printf("Error while update status is_removed by id %s, reason: %v\n", ad.ID, err)
			}

			emails, err := m.service.GetConfirmationService().GetReceivers(ctx, ad.ID)
			if err != nil {
				m.log.Printf("Error while get receivers by id %s, reason: %v\n", ad.ID, err)
			}

			for _, email := range emails {
				go func(innerEmail string) {
					err := m.service.GetEmailService().SendEmailClosedAd(innerEmail, ad.Name, ad.Link)
					if err != nil {
						m.log.Printf("Error while send mail about closed ad, reason: %v\n", err)
					}
				}(email)
			}
		}
	}()

	go func() {
		for {
			ad, ok := <-channelUpdated
			if !ok {
				break
			}
			ctx := context.Background()

			emails, err := m.service.GetConfirmationService().GetReceivers(ctx, ad.ID)
			if err != nil {
				m.log.Printf("Error while get receivers by advert id %s, reason: %v", ad.ID, err)
				continue
			}

			err = m.service.GetAdsService().UpdateAdPrice(ctx, ad.ID, ad.Price)
			if err != nil {
				m.log.Printf("Error while update by advert id %s, reason: %v", ad.ID, err)
				continue
			}

			for _, email := range emails {
				go func(innerEmail string) {
					err := m.service.GetEmailService().SendEmailChangePrice(innerEmail, ad.Name, ad.Price)
					if err != nil {
						m.log.Printf("Error while send mail about changed price by advert id %s, reason: %v\n", ad.ID, err)
					}
				}(email)
			}
		}
	}()

	doneSignal := make(chan bool)
	go func() {
		m.wg.Wait()
		removedEnded := false
		updatedEnded := false
		for {
			if removedEnded && updatedEnded {
				break
			}
			if len(channelUpdated) == 0 {
				updatedEnded = true
			}
			if len(channelRemoved) == 0 {
				removedEnded = true
			}
		}
		close(doneSignal)
	}()

	<-doneSignal

	err = m.service.GetAdsService().UpdateLastChecked(ctx, channelUpdateLastCheck)
	if err != nil {
		m.log.Printf("Error while update last_checked in monitoring, reason: %v", err)
		return err
	}

	m.isRun = false
	return nil
}

func (m *monitoring) Run() {
	var s = gocron.NewScheduler()
	s.Every(10).Seconds().Do(m.executeJob)
	<-s.Start()
}
