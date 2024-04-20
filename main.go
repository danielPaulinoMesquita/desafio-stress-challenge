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
	StatusDistribution map[int]int
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
	fmt.Printf("Distribuição de outros códigos de status HTTP:\n")
	for statusCode, count := range report.StatusDistribution {
		if statusCode != http.StatusOK {
			fmt.Printf("Status %d: %d\n", statusCode, count)
		}
	}
}

func runBenchmark(url string, totalRequests, concurrencyLevel int) Report {
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	semaphore := make(chan struct{}, concurrencyLevel)

	startTime := time.Now()
	statusDistribution := make(map[int]int)

	for i := 0; i < totalRequests; i++ {
		semaphore <- struct{}{}
		go func() {
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Erro ao realizar request para %s: %s\n", url, err.Error())
				return
			}
			defer resp.Body.Close()

			statusCode := resp.StatusCode
			statusDistribution[statusCode]++

			fmt.Printf("Request para %s concluído com status %s\n", url, resp.Status)

			<-semaphore
		}()
	}

	wg.Wait()

	totalTime := time.Since(startTime)
	successfulRequests := totalRequests - statusDistribution[http.StatusInternalServerError]

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
