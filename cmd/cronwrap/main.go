package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourorg/cronwrap/internal/config"
	"github.com/yourorg/cronwrap/internal/history"
	"github.com/yourorg/cronwrap/internal/notifier"
	"github.com/yourorg/cronwrap/internal/runner"
)

func main() {
	cfgPath := flag.String("config", "cronwrap.yaml", "path to config file")
	showHistory := flag.Bool("history", false, "print run history and exit")
	exportCSV := flag.String("export-csv", "", "export history to CSV file")
	limit := flag.Int("limit", 20, "number of records to show in history")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("cronwrap: load config: %v", err)
	}

	store, err := history.NewStore(cfg.History.File)
	if err != nil {
		log.Fatalf("cronwrap: open history store: %v", err)
	}

	if *showHistory {
		if err := history.PrintReport(store, *limit, os.Stdout); err != nil {
			log.Fatalf("cronwrap: print history: %v", err)
		}
		return
	}

	if *exportCSV != "" {
		f, err := os.Create(*exportCSV)
		if err != nil {
			log.Fatalf("cronwrap: create csv: %v", err)
		}
		defer f.Close()
		if err := history.ExportCSV(store, f); err != nil {
			log.Fatalf("cronwrap: export csv: %v", err)
		}
		fmt.Fprintf(os.Stderr, "exported history to %s\n", *exportCSV)
		return
	}

	result, err := runner.Run(runner.Options{
		Command: cfg.Command,
		Args:    cfg.Args,
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	})
	if err != nil {
		log.Fatalf("cronwrap: run: %v", err)
	}

	jobResult := notifier.JobResult{
		JobName:   cfg.JobName,
		Success:   result.ExitCode == 0,
		ExitCode:  result.ExitCode,
		Output:    result.Output,
		Duration:  result.Duration,
		StartedAt: result.StartedAt,
	}

	if err := store.Append(history.Record{
		JobName:   cfg.JobName,
		StartedAt: result.StartedAt,
		Duration:  result.Duration,
		ExitCode:  result.ExitCode,
		Success:   jobResult.Success,
		Output:    result.Output,
	}); err != nil {
		log.Printf("cronwrap: save history: %v", err)
	}

	policy := history.RetentionPolicy{
		MaxRecords: cfg.History.MaxRecords,
		MaxAgeDays: cfg.History.MaxAgeDays,
	}
	if err := history.ApplyToStore(store, policy); err != nil {
		log.Printf("cronwrap: apply retention: %v", err)
	}

	if !jobResult.Success || cfg.NotifyOnSuccess {
		var notifiers []notifier.Notifier
		if cfg.Slack.WebhookURL != "" {
			notifiers = append(notifiers, notifier.NewSlackNotifier(cfg.Slack.WebhookURL))
		}
		if cfg.Email.SMTPHost != "" {
			notifiers = append(notifiers, notifier.NewEmailNotifier(notifier.EmailConfig{
				SMTPHost: cfg.Email.SMTPHost,
				SMTPPort: cfg.Email.SMTPPort,
				Username: cfg.Email.Username,
				Password: cfg.Email.Password,
				From:     cfg.Email.From,
				To:       cfg.Email.To,
			}))
		}
		if cfg.Webhook.URL != "" {
			notifiers = append(notifiers, notifier.NewWebhookNotifier(cfg.Webhook.URL))
		}
		if cfg.PagerDuty.IntegrationKey != "" {
			notifiers = append(notifiers, notifier.NewPagerDutyNotifier(cfg.PagerDuty.IntegrationKey))
		}
		multi := notifier.NewMulti(notifiers...)
		if err := multi.Notify(jobResult); err != nil {
			log.Printf("cronwrap: notify: %v", err)
		}
	}

	if !jobResult.Success {
		os.Exit(1)
	}
}
