// Command cronwrap is a lightweight cron job wrapper that captures stdout/stderr,
// reports failures to Slack or email, and tracks run history.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/yourorg/cronwrap/internal/notifier"
	"github.com/yourorg/cronwrap/internal/runner"
)

const version = "0.1.0"

func main() {
	var (
		slackWebhook  = flag.String("slack-webhook", "", "Slack incoming webhook URL for failure notifications")
		timeoutSecs   = flag.Int("timeout", 0, "Command timeout in seconds (0 = no timeout)")
		jobName       = flag.String("name", "", "Human-readable job name used in notifications")
		notifyOnSuccess = flag.Bool("notify-success", false, "Send a notification even on successful runs")
		showVersion   = flag.Bool("version", false, "Print version and exit")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: cronwrap [options] -- <command> [args...]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  cronwrap --name=backup --slack-webhook=https://... -- /usr/bin/backup.sh --full\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("cronwrap %s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: no command specified")
		flag.Usage()
		os.Exit(2)
	}

	// Derive a job name from the command if not provided.
	name := *jobName
	if name == "" {
		name = strings.Join(args, " ")
		if len(name) > 60 {
			name = name[:60] + "..."
		}
	}

	timeout := time.Duration(*timeoutSecs) * time.Second

	result, err := runner.Run(runner.Options{
		Command: args[0],
		Args:    args[1:],
		Timeout: timeout,
	})
	if err != nil {
		log.Fatalf("cronwrap: failed to run command: %v", err)
	}

	// Always print captured output so cron can log it normally.
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	shouldNotify := !result.Success || *notifyOnSuccess

	if shouldNotify && *slackWebhook != "" {
		n := notifier.NewSlackNotifier(*slackWebhook)
		if notifyErr := n.Notify(notifier.NotifyRequest{
			JobName:  name,
			Success:  result.Success,
			ExitCode: result.ExitCode,
			Stdout:   result.Stdout,
			Stderr:   result.Stderr,
			Duration: result.Duration,
		}); notifyErr != nil {
			log.Printf("cronwrap: slack notification failed: %v", notifyErr)
		}
	}

	if !result.Success {
		os.Exit(result.ExitCode)
	}
}
