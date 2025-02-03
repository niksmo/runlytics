package storage

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

type fileScheme struct {
	Counter map[string]int64   `json:"counter"`
	Gauge   map[string]float64 `json:"gauge"`
}

type ReadUpdateRepository interface {
	UpdateCounterByName(name string, value int64) int64
	UpdateGaugeByName(name string, value float64) float64
	ReadCounter() map[string]int64
	ReadGauge() map[string]float64
}

type FileStorage struct {
	repository ReadUpdateRepository
	interval   time.Duration
	file       *os.File
	restorable bool
	encoder    *json.Encoder
	decoder    *json.Decoder
}

func NewFileStorage(
	repository ReadUpdateRepository,
	interval time.Duration,
	file *os.File,
	restore bool,
) *FileStorage {

	storage := &FileStorage{
		repository: repository,
		interval:   interval,
		file:       file,
		restorable: restore,
		encoder:    json.NewEncoder(file),
	}

	if storage.restorable {
		storage.decoder = json.NewDecoder(file)
	}

	return storage
}

func (s *FileStorage) Run() {
	s.restore()

	if !s.isSync() {
		go s.intervalSave()
	}

	go s.interceptSigInt()
}

func (s *FileStorage) restore() {

	if !s.restorable {
		s.file.Truncate(0)
		return
	}

	var data fileScheme
	err := s.decoder.Decode(&data)
	if errors.Is(err, io.EOF) {
		logger.Log.Info("File storage is empty")
		return
	}

	if err != nil {
		logger.Log.Error("JSON decode", zap.Error(err))
		return
	}

	for name, value := range data.Gauge {
		s.repository.UpdateGaugeByName(name, value)
	}

	for name, value := range data.Counter {
		s.repository.UpdateCounterByName(name, value)
	}
	logger.Log.Info(
		"Restore metrics",
		zap.Int("count", len(data.Counter)+len(data.Gauge)),
	)
}

func (s *FileStorage) isSync() bool {
	return s.interval == 0
}

func (s *FileStorage) save() {
	data := fileScheme{
		Gauge:   s.repository.ReadGauge(),
		Counter: s.repository.ReadCounter(),
	}

	s.file.Seek(0, 0)
	if err := s.encoder.Encode(data); err != nil {
		logger.Log.Error("JSON encode", zap.Error(err))
	}
	logger.Log.Debug("Save metrics to file")
}

func (s *FileStorage) intervalSave() {
	for {
		time.Sleep(s.interval)
		s.save()
	}
}

func (s *FileStorage) close() {
	s.save()
	if err := s.file.Close(); err != nil {
		logger.Log.Error("Close storage file", zap.Error(err))
	}
}

func (s *FileStorage) interceptSigInt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	sig := <-c
	logger.Log.Debug("Got signal", zap.String("signal", sig.String()))
	s.close()
	os.Exit(0)
}

func (s *FileStorage) UpdateCounterByName(name string, value int64) int64 {
	ret := s.repository.UpdateCounterByName(name, value)

	if s.isSync() {
		s.save()
	}

	return ret
}
func (s *FileStorage) UpdateGaugeByName(name string, value float64) float64 {
	ret := s.repository.UpdateGaugeByName(name, value)

	if s.isSync() {
		s.save()
	}

	return ret
}
