package bot

import (
	"context"
	"sync"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/bot/commands"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

type CommandJob struct {
	Session *discordgo.Session
	Message *discordgo.MessageCreate
	Command string
	Args    []string
}

type CommandProcessor struct {
	jobQueue    chan *CommandJob
	workerCount int
	logger      *logger.Logger
	vm          *verification.VerificationManager
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewCommandProcessor(workerCount int, log *logger.Logger, vm *verification.VerificationManager) *CommandProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &CommandProcessor{
		jobQueue:    make(chan *CommandJob, 100),
		workerCount: workerCount,
		logger:      log,
		vm:          vm,
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

	handler, ok := commands.Registry[job.Command]
	if !ok {
		cp.logger.Debug("Unknown command", "command", job.Command, "user", job.Message.Author.Username)
		return
	}

	done := make(chan error, 1)
	go func() {
		done <- handler(job.Session, job.Message, job.Args, cp.logger, cp.vm)
	}()

	select {
	case err := <-done:
		if err != nil {
			cp.logger.Error("Command execution failed", "command", job.Command, "error", err)
			job.Session.ChannelMessageSend(job.Message.ChannelID, "❌ An error occurred while processing your command.")
		}
	case <-time.After(20 * time.Second):
		cp.logger.Warn("Command timed out", "command", job.Command, "user", job.Message.Author.Username)
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
	case <-time.After(10 * time.Second):
		cp.logger.Warn("Command processor shutdown timed out")
	}
}

func (cp *CommandProcessor) Len() int {
	return len(cp.jobQueue)
}
