package workerpool

import (
	"github.com/rs/zerolog/log"
	"sync"
)

type Job func() error

type Pool struct {
	workerCount int
	jobs        chan Job
	wg          sync.WaitGroup
}

func New(workerCount int) *Pool {
	return &Pool{
		workerCount: workerCount,
		jobs:        make(chan Job, workerCount),
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for job := range p.jobs {
		err := job()
		if err != nil {
			log.Fatal().Err(err).Msgf("Worker %d: Error executing job", id)
		}
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *Pool) AddJob(job Job) {
	p.jobs <- job
}

func (p *Pool) Close() {
	close(p.jobs)
	p.wg.Wait()
}
