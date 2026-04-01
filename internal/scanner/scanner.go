package scanner

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	Host    string
	Start   int
	End     int
	Workers int
	Timeout time.Duration
}

type Scanner struct {
	cfg Config
}

func New(cfg Config) *Scanner {
	return &Scanner{cfg: cfg}
}

func (s *Scanner) Scan(ctx context.Context) ([]Result, error) {
	jobs := make(chan int, s.cfg.Workers*2)
	resultsCh := make(chan Result, s.cfg.Workers*2)

	var wg sync.WaitGroup

	for i := 0; i < s.cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.worker(ctx, jobs, resultsCh)
		}()
	}

	go func() {
		defer close(jobs)

		for port := s.cfg.Start; port <= s.cfg.End; port++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- port:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]Result, 0, 64)

	for result := range resultsCh {
		results = append(results, result)
	}

	if err := ctx.Err(); err != nil {
		return results, err
	}

	return results, nil
}

func (s *Scanner) worker(ctx context.Context, jobs <-chan int, results chan<- Result) {
	for {
		select {
		case <-ctx.Done():
			return
		case port, ok := <-jobs:
			if !ok {
				return
			}

			open := s.isOpen(ctx, port)
			if !open {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case results <- Result{Port: port}:
			}
		}
	}
}

func (s *Scanner) isOpen(ctx context.Context, port int) bool {
	address := net.JoinHostPort(s.cfg.Host, strconv.Itoa(port))

	dialCtx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(dialCtx, "tcp", address)
	if err != nil {
		return false
	}

	_ = conn.Close()
	return true
}

func (s *Scanner) String() string {
	return fmt.Sprintf(
		"Scanner{host=%s,start=%d,end=%d,workers=%d,timeout=%s}",
		s.cfg.Host,
		s.cfg.Start,
		s.cfg.End,
		s.cfg.Workers,
		s.cfg.Timeout,
	)
}
