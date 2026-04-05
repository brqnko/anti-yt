package scheduler

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/robfig/cron/v3"
)

type Job interface {
	Run()
}

type Service interface {
	AddFunc(spec string, task Job) (err error)
	Stop()
}

var _ Service = (*serviceImpl)(nil)

type serviceImpl struct {
	c *cron.Cron
}

func (s *serviceImpl) AddFunc(spec string, task Job) (err error) {
	defer util.Wrap(&err, "scheduler.(*serviceImpl).AddFunc: %s", spec)

	_, err = s.c.AddFunc(spec, task.Run)

	return nil
}

func (s *serviceImpl) Stop() {
	s.c.Stop()
}

func NewService() Service {
	c := cron.New(cron.WithLocation(time.UTC), cron.WithLogger(cron.DefaultLogger))
	c.Start()

	return &serviceImpl{c: c}
}
