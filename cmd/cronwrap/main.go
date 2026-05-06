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
		records, _ := store.ReadAll()
		if err := history.ExportCSV(records, f); err != nil {
			log.Fatalf("cronwrap: export csv: %v", err)
		}
		fmt.Fprintf(os.Stdout, "exported %d records to %s\n", len(records), *exportCSV)
		return
	}

	result, err := runner.Run(cfg)
	if err != nil {
		log.Fatalf("cronwrap: run: %v", err)
	}

	rec := history.Record{
		JobName:   cfg.JobName,
		StartedAt: time.Now().Add(-result.Duration),
		Duration:  result.Duration,
		Success:   result.Success,
		ExitCode:  result.ExitCode,
		Output:    result.Output,
	}
	if err := store.Append(rec); err != nil {
		log.Printf("cronwrap: save history: %v", err)
	}

	policy := history.RetentionPolicy{
		MaxRecords: cfg.History.MaxRecords,
		MaxAgeDays: cfg.History.MaxAgeDays,
	}
	if err := history.ApplyToStore(store, policy); err != nil {
		log.Printf("cronwrap: apply retention: %v", err)
	}

	if !result.Success || cfg.NotifyOnSuccess {
		var notifiers []notifier.Notifier
		if cfg.Slack.WebhookURL != "" {
			notifiers = append(notifiers, notifier.NewSlackNotifier(cfg.Slack.WebhookURL))
		}
		if cfg.Webhook.URL != "" {
			notifiers = append(notifiers, notifier.NewWebhookNotifier(cfg.Webhook.URL))
		}
		if cfg.PagerDuty.RoutingKey != "" {
			notifiers = append(notifiers, notifier.NewPagerDutyNotifier(cfg.PagerDuty.RoutingKey))
		}
		if cfg.OpsGenie.APIKey != "" {
			notifiers = append(notifiers, notifier.NewOpsGenieNotifier(cfg.OpsGenie.APIKey, cfg.OpsGenie.Team))
		}
		if cfg.Email.SMTPHost != "" {
			notifiers = append(notifiers, notifier.NewEmailNotifier(
				cfg.Email.SMTPHost, cfg.Email.SMTPPort,
				cfg.Email.Username, cfg.Email.Password,
				cfg.Email.From, cfg.Email.To,
			))
		}
		multi := notifier.NewMulti(notifiers...)
		if err := multi.Notify(result); err != nil {
			log.Printf("cronwrap: notify: %v", err)
		}
	}

	if !result.Success {
		os.Exit(result.ExitCode)
	}
}
