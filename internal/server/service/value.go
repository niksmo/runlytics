package service

type ReadService struct {
	repository ReadByNameRepository
}

type ReadByNameRepository interface {
	ReadCounterByName(name string) (int64, error)
	ReadGaugeByName(name string) (float64, error)
}

func NewReadService(repository ReadByNameRepository) *ReadService {
	return &ReadService{repository}
}
