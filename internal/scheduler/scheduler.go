package scheduler

import (
	"log"

	"schools-be/internal/config"
	"schools-be/internal/service"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron          *cron.Cron
	schoolService *service.SchoolService
	config        *config.Config
}

func New(cfg *config.Config, schoolService *service.SchoolService) *Scheduler {
	return &Scheduler{
		cron:          cron.New(),
		schoolService: schoolService,
		config:        cfg,
	}
}

func (s *Scheduler) Start() {
	// Schedule school data refresh
	_, err := s.cron.AddFunc(s.config.FetchSchedule, func() {
		log.Println("Running scheduled school data refresh...")
		if err := s.schoolService.RefreshSchoolsData(); err != nil {
			log.Printf("Scheduled refresh failed: %v", err)
		}
	})
	if err != nil {
		log.Printf("Failed to schedule job: %v", err)
	}

	// Add more scheduled jobs here as needed
	// Example:
	// s.cron.AddFunc("0 3 * * *", s.cleanupOldData)

	s.cron.Start()
	log.Printf("Scheduler started with schedule: %s", s.config.FetchSchedule)
}

func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.cron.Stop()
}

// Add more scheduled job functions here
// func (s *Scheduler) cleanupOldData() {
//     log.Println("Running cleanup job...")
//     // Implementation
// }
