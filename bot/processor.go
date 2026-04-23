package bot

import (
	"context"
	"sync"

	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/bwmarrin/discordgo"
)

type CommandJob struct {
	Session *discordgo.Session
	Message *discordgo.MessageCreate
	Command string
	Args    []string
	Ctx     context.Context
}

type CommandProcessor struct {
	jobQueue    chan *CommandJob
	workerCount int
	logger      *logger.Logger
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewCommandProcessor(workerCount int, log *logger.Logger) *CommandProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &CommandProcessor{
		jobQueue:    make(chan *CommandJob, 100),
		workerCount: workerCount,
		logger:      log,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (cp *CommandProcessor) Start() {
	cp.logger.Info("Starting command processor", "workers", cp.workerCount)

	for i := 0; i < cp.workerCount; i++ {
		cp.wg.Add(1)
		go cp.worker(i)
	}
}

func (cp *CommandProcessor) worker(id int) {
	defer cp.wg.Done()

	cp.logger.Debug("Worker started", "worker_id", id)

	for {
		select {
		case <-cp.ctx.Done():
			cp.logger.Debug("Worker shutting down", "worker_id", id)
			return
		case job := <-cp.jobQueue:
			if job == nil {
				return
			}
			cp.processJob(job, id)
		}
	}
}

func (cp *CommandProcessor) processJob(job *CommandJob, workerID int) {
	cp.logger.Debug("Processing command", "command", job.Command, "worker_id", workerID, "user", job.Message.Author.Username)

	jobCtx, cancel := context.WithTimeout(job.Ctx, 0)
	defer cancel()

	select {
	case <-jobCtx.Done():
		cp.logger.Warn("Command job cancelled before processing", "command", job.Command)
		return
	default:
	}
}

func (cp *CommandProcessor) Submit(job *CommandJob) error {
	select {
	case <-cp.ctx.Done():
		cp.logger.Warn("Cannot submit job - processor is shutting down", "command", job.Command)
		return context.Canceled
	case cp.jobQueue <- job:
		return nil
	}
}

func (cp *CommandProcessor) Shutdown() {
	cp.logger.Info("Shutting down command processor")
	cp.cancel()

	done := make(chan struct{})
	go func() {
		cp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		cp.logger.Info("Command processor shut down successfully")
	case <-context.Background().Done():
		cp.logger.Warn("Command processor shutdown timed out")
	}
}

func (cp *CommandProcessor) Len() int {
	return len(cp.jobQueue)
}
