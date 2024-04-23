package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"sync"
	"time"
)

type Report struct {
	TotalTime          time.Duration
	TotalRequests      int
	SuccessfulRequests int
	StatusDistribution int
}

func runReportCmd(cmd *cobra.Command, args []string) {
	url, _ := cmd.Flags().GetString("url")
	totalRequests, _ := cmd.Flags().GetInt("requests")
	concurrencyLevel, _ := cmd.Flags().GetInt("concurrency")

	report := runBenchmark(url, totalRequests, concurrencyLevel)

	fmt.Printf("Relatório:\n")
	fmt.Printf("Tempo total gasto: %s\n", report.TotalTime)
	fmt.Printf("Quantidade total de requests: %d\n", report.TotalRequests)
	fmt.Printf("Quantidade de requests com status HTTP 200: %d\n", report.SuccessfulRequests)
	fmt.Printf("Distribuição de outros códigos de status HTTP: %d\n", report.StatusDistribution)
}

func runBenchmark(url string, totalRequests, concurrencyLevel int) Report {
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	semaphore := make(chan struct{}, concurrencyLevel)

	startTime := time.Now()

	var mu sync.Mutex
	statusDistribution := 0

	for i := 0; i < totalRequests; i++ {
		semaphore <- struct{}{}
		go func() {
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Erro ao realizar request para %s: %s\n", url, err.Error())
				mu.Lock()
				statusDistribution++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			statusCode := resp.StatusCode
			if statusCode < 200 || statusCode >= 300 {
				mu.Lock()
				statusDistribution++
				mu.Unlock()
			}

			fmt.Printf("Request para %s concluído com status %s\n", url, resp.Status)

			<-semaphore
		}()
	}

	wg.Wait()

	totalTime := time.Since(startTime)

	successfulRequests := totalRequests - statusDistribution

	return Report{
		TotalTime:          totalTime,
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		StatusDistribution: statusDistribution,
	}
}

func main() {
	var (
		url              string
		totalRequests    int
		concurrencyLevel int
	)

	benchmarkCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate report",
		Run:   runReportCmd,
	}

	benchmarkCmd.PersistentFlags().StringVar(&url, "url", "", "URL to benchmark")
	benchmarkCmd.PersistentFlags().IntVar(&totalRequests, "requests", 0, "Total number of requests")
	benchmarkCmd.PersistentFlags().IntVar(&concurrencyLevel, "concurrency", 0, "Concurrency level")

	if err := benchmarkCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
