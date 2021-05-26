package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bottlepay/portfolio-data/model"
	"github.com/bottlepay/portfolio-data/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

// dataCmd represents the data command
var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Simulate Custodian data",
	Long:  `Simulate Custodian data`,
	RunE:  dataRunner,
}

func init() {
	rootCmd.AddCommand(dataCmd)

	dataCmd.PersistentFlags().StringP("state", "s", "state.json", "file to save state to")
	dataCmd.PersistentFlags().StringP("listen", "l", "0.0.0.0:9999", "the address to listen on")
	dataCmd.PersistentFlags().IntP("timer", "t", 1, "the frequency (in seconds) with which to generate events. 0 disables automatic generation")

	rand.Seed(time.Now().UnixNano())
}

func dataRunner(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	stateFile, err := flags.GetString("state")
	if err != nil {
		return err
	}

	listenAddr, err := flags.GetString("listen")
	if err != nil {
		return err
	}

	eventFrequency, err := flags.GetInt("timer")
	if err != nil {
		return err
	}

	// Load the data store
	store, err := store.NewStore(stateFile)
	if err != nil {
		return err
	}

	// Setup the store. This will generate initial data if the store is empty
	setupStore(store)

	// Setup the HTTP server, and listen on the desired address
	server := NewServerContext(store, listenAddr)
	listenErr := make(chan error)
	go func() {
		fmt.Println("listening for HTTP traffic on", listenAddr)
		err := server.ListenAndServe()
		listenErr <- err
	}()

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Setup a ticker we'll use to generate new events, if that functionality is enabled
	var tickerChannel <-chan time.Time
	if eventFrequency > 0 {
		ticker := time.NewTicker(time.Second * time.Duration(eventFrequency))
		tickerChannel = ticker.C
	} else {
		// Otherwise just create a blank channel
		tickerChannel = make(chan time.Time)
	}

	// Setup the main event loop
	for {
		select {
		case err := <-listenErr:
			// Return any HTTP server errors, unless it's the shutdown error
			if err == http.ErrServerClosed {
				return nil
			}
			return err
		case <-c:
			// On interrupt, shutdown the server
			fmt.Println("triggering shutdown from interrupt")
			server.Shutdown(context.Background())
		case <-tickerChannel:
			// Every time the ticker ticks, add a new random event
			err := store.AddRandomEvent()
			if err != nil {
				return err
			}

			err = store.Snapshot()
			if err != nil {
				return err
			}
		}
	}
}

func setupStore(store *store.Store) {
	// Setup the initial data if the store is empty
	if !store.IsEmpty() {
		return
	}

	store.AddCustodian(
		// A BTC wallet
		&model.Custodian{
			Assets: []*model.Asset{
				{
					Code:    "BTC",
					Balance: decimal.RequireFromString("10.00000001"),
				},
			},
		},
		// An exchange
		&model.Custodian{
			Assets: []*model.Asset{
				{
					Code:    "BTC",
					Balance: decimal.RequireFromString("10.00000001"),
				},
				{
					Code:    "GBP",
					Balance: decimal.RequireFromString("100000.01"),
				},
			},
		},
		// An exchange
		&model.Custodian{
			Assets: []*model.Asset{
				{
					Code:    "BTC",
					Balance: decimal.RequireFromString("10.00000001"),
				},
				{
					Code:    "GBP",
					Balance: decimal.RequireFromString("100000.01"),
				},
			},
		},
		// An exchange
		&model.Custodian{
			Assets: []*model.Asset{
				{
					Code:    "BTC",
					Balance: decimal.RequireFromString("10.00000001"),
				},
				{
					Code:    "GBP",
					Balance: decimal.RequireFromString("100000.01"),
				},
			},
		},
	)

	for i := 0; i < 100; i++ {
		store.AddRandomEvent()
	}

	store.Snapshot()
}

type ServerContext struct {
	store  *store.Store
	server *http.Server
}

func NewServerContext(store *store.Store, listenAddr string) *ServerContext {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	server := &http.Server{
		Handler: cors.AllowAll().Handler(r),
		Addr:    listenAddr,
	}

	serverCtx := &ServerContext{
		store:  store,
		server: server,
	}

	r.Get("/custodian/{id}", serverCtx.HandleGetCustodian)
	r.Get("/generate", serverCtx.HandleGenerate)

	return serverCtx
}

func (s *ServerContext) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *ServerContext) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *ServerContext) HandleGetCustodian(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if id < 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	custodian := s.store.GetCustodian(int32(id))
	if custodian == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Add("content-type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(custodian)
}

func (s *ServerContext) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.ParseInt(r.URL.Query().Get("count"), 10, 64)
	if count < 0 {
		count = 1
	}
	if count > 1000 {
		count = 1000
	}

	for i := int64(0); i < count; i++ {
		err := s.store.AddRandomEvent()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)

			return
		}
	}

	err := s.store.Snapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)

		return
	}
}
