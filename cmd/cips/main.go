package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"cips/internal/domain"
	"cips/internal/engine"
	"cips/internal/exporter"
	"cips/internal/ingestion"
	"cips/internal/pipeline"
	"cips/internal/probe"
	"cips/internal/server"
	"cips/web"
)

var (
	port         int
	headless     bool
	concurrency  int
	pingTimeout  int
	probeTimeout int
	inputFile    string
	subURL       string
	outputSub    string
	outputJSON   string
)

var rootCmd = &cobra.Command{
	Use:   "cips",
	Short: "Clean IP Scanner — Enterprise High-Throughput Proxy AI Validator",
	Long: `CIPS (Clean IP Scanner) is an enterprise-grade high-throughput proxy validator
that evaluates proxies for clean IP reputation against Google Gemini, Google AI Studio,
and enriches active endpoints with exit IP and ISP intelligence from ipleak.`,
	Run: runApplication,
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 19090, "Default port for embedded web dashboard")
	rootCmd.PersistentFlags().BoolVar(&headless, "headless", false, "Run in headless mode without auto-opening web browser")
	rootCmd.PersistentFlags().BoolVar(&headless, "no-browser", false, "Alias for --headless")
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", 20, "Number of concurrent scanning workers")
	rootCmd.PersistentFlags().IntVar(&pingTimeout, "ping-timeout", 3000, "Stage 1 ping reachability timeout in milliseconds")
	rootCmd.PersistentFlags().IntVar(&probeTimeout, "probe-timeout", 8000, "Stage 2 service probing timeout in milliseconds")
	rootCmd.PersistentFlags().StringVarP(&inputFile, "file", "f", "", "Optional path to proxy list file for direct CLI scanning")
	rootCmd.PersistentFlags().StringVarP(&subURL, "sub", "s", "", "Optional remote subscription URL for direct CLI scanning")
	rootCmd.PersistentFlags().StringVarP(&outputSub, "out", "o", "clean_sub.txt", "Path for clean exported subscription file")
	rootCmd.PersistentFlags().StringVar(&outputJSON, "json", "", "Optional path to save full session telemetry JSON")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printBanner() {
	banner := `
  ██████╗██╗██████╗ ███████╗
 ██╔════╝██║██╔══██╗██╔════╝
 ██║     ██║██████╔╝███████╗  Clean IP Scanner v1.0.0
 ██║     ██║██╔═══╝ ╚════██║  Enterprise High-Throughput Proxy AI Validator
 ╚██████╗██║██║     ███████║  [Google Gemini | AI Studio | IPLeak]
  ╚═════╝╚═╝╚═╝     ╚══════╝
`
	fmt.Print(banner)
}

func runApplication(cmd *cobra.Command, args []string) {
	printBanner()

	// 1. Ensure Xray Core
	fmt.Println("[Core] Checking Xray proxy engine...")
	xrayPath, err := engine.EnsureXrayBinary(func(percent float64, message string) {
		fmt.Printf("\r[Core] %s (%.0f%%)", message, percent)
	})
	if err != nil {
		fmt.Printf("\n[Error] Failed to prepare Xray engine: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n[Core] Xray engine active at: %s\n", xrayPath)

	// 2. Initialize Engine & Probes
	xrayEngine := engine.NewXrayEngine(xrayPath)
	defer xrayEngine.Shutdown()

	probes := []domain.IProbe{
		probe.NewGeminiProbe(probeTimeout),
		probe.NewAIStudioProbe(probeTimeout),
		probe.NewIPLeakProbe(probeTimeout),
	}

	// 3. Setup Signal Handler
	ctx := pipeline.SetupSignalHandler(func() {
		fmt.Println("\n[Interrupt] Shutdown signal received. Releasing resources...")
		_ = xrayEngine.Shutdown()
	})

	// 4. Check if direct CLI mode requested
	if inputFile != "" || subURL != "" {
		runCLIMode(ctx, xrayEngine, probes)
		return
	}

	// 5. Interactive Web Dashboard Mode
	callbacks := pipeline.OrchestratorCallbacks{}
	orch := pipeline.NewOrchestrator(xrayEngine, probes, concurrency, pingTimeout, probeTimeout, callbacks)

	srv := server.NewServer(orch, web.Assets)

	// Wire live broadcaster callbacks
	orchCallbacks := pipeline.OrchestratorCallbacks{
		OnStatsUpdate: func(stats domain.ScanSessionStats) {
			srv.Broadcaster().BroadcastStats(stats)
		},
		OnItemUpdate: func(report *domain.ProxyEvaluationReport) {
			srv.Context().UpdateReport(report)
			srv.Broadcaster().BroadcastItemUpdate(report)
		},
	}
	// Re-construct with wired callbacks
	orch = pipeline.NewOrchestrator(xrayEngine, probes, concurrency, pingTimeout, probeTimeout, orchCallbacks)
	srv = server.NewServer(orch, web.Assets)

	boundPort, err := srv.ListenWithFallback(port)
	if err != nil {
		fmt.Printf("[Error] Failed to bind local web dashboard: %v\n", err)
		os.Exit(1)
	}

	dashboardURL := fmt.Sprintf("http://127.0.0.1:%d", boundPort)
	fmt.Printf("[Dashboard] Live web interface serving at: %s\n", dashboardURL)

	if !headless {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = server.OpenBrowser(dashboardURL)
		}()
	}

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[Server Error] %v\n", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
	fmt.Println("[Exit] Clean shutdown complete.")
}

func runCLIMode(ctx context.Context, xrayEngine *engine.XrayEngine, probes []domain.IProbe) {
	fmt.Println("[Mode] Running direct CLI batch scan...")
	var candidates []domain.ProxyConfig
	var totalParsed int

	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("[Error] Failed to read input file %s: %v\n", inputFile, err)
			os.Exit(1)
		}
		candidates, totalParsed = ingestion.ExtractAndDeduplicate(string(data))
	} else if subURL != "" {
		fmt.Printf("[Ingest] Fetching remote subscription: %s\n", subURL)
		cfgs, total, err := ingestion.FetchSubscription(ctx, subURL)
		if err != nil {
			fmt.Printf("[Error] Failed to fetch subscription: %v\n", err)
			os.Exit(1)
		}
		candidates = cfgs
		totalParsed = total
	}

	fmt.Printf("[Ingest] Parsed %d total configurations (%d unique candidates)\n", totalParsed, len(candidates))
	if len(candidates) == 0 {
		fmt.Println("[Warning] No valid proxy configurations found to scan.")
		return
	}

	callbacks := pipeline.OrchestratorCallbacks{
		OnStatsUpdate: func(stats domain.ScanSessionStats) {
			fmt.Printf("\r[Progress] Alive: %d | Dead: %d | AI Clean: %d / %d",
				stats.Stage1Alive, stats.Stage1Dead, stats.Stage2Clean, stats.TotalIngested)
		},
	}

	orch := pipeline.NewOrchestrator(xrayEngine, probes, concurrency, pingTimeout, probeTimeout, callbacks)
	reports, err := orch.Execute(ctx, candidates)
	if err != nil && err != context.Canceled {
		fmt.Printf("\n[Error] Pipeline execution error: %v\n", err)
		return
	}

	fmt.Println("\n\n[Results] Processing complete. Generating output...")

	// Clean subscription export
	subText := exporter.GenerateCleanSubscription(reports, "")
	if err := os.WriteFile(outputSub, []byte(subText), 0644); err != nil {
		fmt.Printf("[Error] Failed to write clean subscription: %v\n", err)
	} else {
		fmt.Printf("[Export] Clean subscription written to: %s\n", outputSub)
	}

	// JSON telemetry export
	if outputJSON != "" {
		reportMap := make(map[string]*domain.ProxyEvaluationReport)
		for _, r := range reports {
			reportMap[r.Config.ID] = r
		}
		stats := domain.ScanSessionStats{
			TotalIngested: totalParsed,
			Deduplicated:  len(candidates),
		}
		jsonPath, err := exporter.WriteSessionLog(stats, reportMap, "")
		if err != nil {
			fmt.Printf("[Error] Failed to write JSON telemetry: %v\n", err)
		} else {
			fmt.Printf("[Export] Full JSON telemetry saved to: %s\n", jsonPath)
		}
	}
}
