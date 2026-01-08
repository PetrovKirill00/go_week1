package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PetrovKirill00/go_week1/cmd/internal/wordstat"
)

func main() {
	addr := flag.String("addr", ":8080", "main listener address")
	dev := flag.Bool("dev", false, "dev mode: disable read/write timeouts")
	readTO := flag.Duration("read-timeout", 10*time.Second, "server read timeout")
	writeTO := flag.Duration("write-timeout", 10*time.Second, "server write timeout")
	shutdownTO := flag.Duration("shutdown-timeout", 10*time.Second, "graceful shutdown timeout")
	enablePprof := flag.Bool("pprof", false, "enable pprof server")
	pprofAddr := flag.String("pprof-addr", "127.0.0.1:6060", "pprof listen address")
	maxBody := flag.Int64("max-body", wordstat.DefaultHTTPConfig.MaxBodyBytes, "number of bytes allowed in POST body")
	flag.Parse()

	rt, wt := *readTO, *writeTO
	if *dev {
		rt, wt = 0, 0
	}

	mainHandler := wordstat.NewHTTPMuxWithConfig(wordstat.HTTPConfig{
		MaxBodyBytes: *maxBody,
	})
	mainSrv := http.Server{
		Addr:              *addr,
		Handler:           mainHandler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,

		ReadTimeout:  rt,
		WriteTimeout: wt,
		// для Dev/тестов со slow client
		// ReadTimeout:  0,
		// WriteTimeout: 0,
	}

	// Контекст, который отменился при ctrl+c / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCap := 1
	if *enablePprof {
		errCap = 2
	}
	errCh := make(chan error, errCap)

	go func() {
		log.Println("main listening on", *addr)
		errCh <- mainSrv.ListenAndServe()
	}()

	var pprofSrv *http.Server
	if *enablePprof {
		pprofMux := http.NewServeMux()
		pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
		pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		pprofSrv = &http.Server{
			Addr:              *pprofAddr,
			Handler:           pprofMux,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       60 * time.Second,
		}
		go func() {
			log.Printf("pprof listening on http://%s/debug/pprof", *pprofAddr)
			errCh <- pprofSrv.ListenAndServe()
		}()
	}

	select {
	case <-ctx.Done():
		log.Printf("shutdown signal received")
	case err := <-errCh:
		// Если сервер упал сам по себе, то это реальная ошибка
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
		// Если ErrServerClosed - значит, кто-то уже вызвал Shutdown/Close
	}

	// Даём текущим запросам закончиться
	shutdownCtx, cancel := context.WithTimeout(context.Background(), *shutdownTO)
	defer cancel()

	if err := mainSrv.Shutdown(shutdownCtx); err != nil {
		// Если не уложились по таймеру, то выходим жестко
		log.Printf("graceful shutdown failed: %v; forcing close", err)
		_ = mainSrv.Close()
	}

	if pprofSrv != nil {
		if err := pprofSrv.Shutdown(shutdownCtx); err != nil {
			// Если не уложились по таймеру, то выходим жестко
			log.Printf("graceful shutdown failed: %v; forcing close", err)
			_ = pprofSrv.Close()
		}
	}

	log.Printf("servers stopped")

}
