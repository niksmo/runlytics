package service

type UpdateService struct {
	repository UpdateRepository
}

type UpdateRepository interface {
	UpdateCounterByName(name string, value int64)
	UpdateGaugeByName(name string, value float64)
}

func NewUpdateService(repository UpdateRepository) *UpdateService {
	return &UpdateService{repository}
}

func (service *UpdateService) Update() {}
