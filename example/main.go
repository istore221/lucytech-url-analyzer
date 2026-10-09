package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	analyzer "github.com/istore221/lucytech-url-analyzer"
)

type AnalysisResult struct {
	URL      string
	Analysis *analyzer.PageAnalysis
	Err      error
}

func main() {
	urls := []string{
		//"https://www.york.ac.uk/teaching/cws/wws/webpage1.html",
		//"https://toscrape.com/",
		//"https://www.scrapethissite.com/pages/",
		"https://www.scrapethissite.com/login/",
		//"https://golang.org",
		//"https://github.com",
		//"https://example.com",
		//"https://this-will-fail-123.com",
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	engine := analyzer.New(
		analyzer.WithWorkers(5),
	)

	var wg sync.WaitGroup
	results := make(chan AnalysisResult, len(urls))

	fmt.Printf("Starting concurrent analysis of %d URLs...\n", len(urls))
	fmt.Println("Press Ctrl+C at any time to cancel.")
	startTime := time.Now()

	for _, u := range urls {
		wg.Add(1)

		go func(targetURL string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(rootCtx, 20*time.Second)
			defer cancel()

			analysis, err := engine.Analyze(ctx, targetURL)

			select {
			case results <- AnalysisResult{
				URL:      targetURL,
				Analysis: analysis,
				Err:      err,
			}:
			case <-rootCtx.Done():
				return
			}
		}(u)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	doneConsuming := make(chan struct{})

	go func() {
		for res := range results {
			printResult(res)
			fmt.Println("---------------------------------------------------")
		}
		close(doneConsuming)
	}()

	select {
	case <-doneConsuming:

		fmt.Printf("Finished all jobs in %v\n", time.Since(startTime))

	case <-rootCtx.Done():

		fmt.Println("\n⚠️  Interrupt received (Ctrl+C). Canceling remaining operations...")

		wg.Wait()
		fmt.Printf("Canceled cleanly in %v\n", time.Since(startTime))
	}
}

func printResult(res AnalysisResult) {
	if res.Err != nil {
		fmt.Printf("❌ [FAILED] %s\n", res.URL)
		switch {
		case errors.Is(res.Err, context.Canceled):
			fmt.Println("   Error: Operation canceled by user")
		case errors.Is(res.Err, context.DeadlineExceeded):
			fmt.Println("   Error: Operation timed out")
		default:
			fmt.Printf("   Error: %v\n", res.Err)
		}
		return
	}

	a := res.Analysis
	fmt.Printf("✅ [SUCCESS] %s\n", a.URL)
	fmt.Printf("   ├─ Title:          %s\n", a.Title)
	fmt.Printf("   ├─ HTML Version:   %s\n", a.HTMLVersion)
	fmt.Printf("   └─ Has Login Form: %t\n\n", a.HasLoginForm)

	if len(a.Headings) > 0 {
		fmt.Println("   [Headings Breakdown]")
		for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
			if count, exists := a.Headings[tag]; exists {
				fmt.Printf("     • %s: %d\n", strings.ToUpper(tag), count)
			}
		}

		for tag, count := range a.Headings {
			lower := strings.ToLower(tag)
			if lower < "h1" || lower > "h6" {
				fmt.Printf("     • %s: %d\n", tag, count)
			}
		}
		fmt.Println()
	}

	fmt.Printf("   [Internal Links - %d]\n", len(a.InternalLinks))
	if len(a.InternalLinks) == 0 {
		fmt.Println("     └─ (none)")
	} else {
		for _, link := range a.InternalLinks {
			fmt.Printf("     • %s\n", link)
		}
	}
	fmt.Println()

	fmt.Printf("   [External Links - %d]\n", len(a.ExternalLinks))
	if len(a.ExternalLinks) == 0 {
		fmt.Println("     └─ (none)")
	} else {
		for _, link := range a.ExternalLinks {
			fmt.Printf("     • %s\n", link)
		}
	}
	fmt.Println()

	fmt.Printf("   [Dead Links - %d]\n", len(a.InaccessibleLinks))
	if len(a.InaccessibleLinks) == 0 {
		fmt.Println("     └─ (none)")
	} else {
		for _, deadLink := range a.InaccessibleLinks {
			fmt.Printf("     • %s\n", deadLink)
		}
	}
}
