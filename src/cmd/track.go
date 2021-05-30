package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bottlepay/portfolio-data/model"
	"github.com/bottlepay/portfolio-data/service"
	"github.com/bottlepay/portfolio-data/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
)

const (
	USERCONTEXT = "user"
)

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Start the tracker service",
	Long:  `The tracker service tracks the portfolios of users`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()

		userStore := store.NewFakeUserStore()
		userStore.Populate()

		url, err := flags.GetString("custodian")
		if err != nil {
			panic("--custodian=URL is required")
		}
		custSvc := service.NewCustodianSvc(url)

		r := chi.NewRouter()
		r.Use(middleware.Logger)
		r.Route("/user/{id}", func(r chi.Router) {
			r.Use(handleUserCtx(userStore))
			r.Get("/", handleUserRoute())
			r.Get("/holdings", handleHoldingsRoute(custSvc))
			r.Route("/custodian/{custId}", func(r chi.Router) {
				r.Get("/transactions", handleTransactionsRoute(custSvc))
			})
		})

		listenAddr, err := flags.GetString("listen")
		if err != nil {
			return err
		}

		server := &http.Server{
			Handler: cors.AllowAll().Handler(r),
			Addr:    listenAddr,
		}
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

		for {
			select {
			case err := <-listenErr:
				if err == http.ErrServerClosed {
					return nil
				}
				return err
			case <-c:
				fmt.Println("triggering shutdown from interrupt")
				server.Shutdown(context.Background())
			}
		}
	},
}

// /user/{id} -> id in USERCONTEXT context value
func handleUserCtx(s store.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			sid := chi.URLParam(r, "id")
			id, err := strconv.Atoi(sid)
			if err != nil {
				http.Error(rw, "invalid id", http.StatusNotFound)
				return
			}
			// I would add JWT token verification logic here
			// the ID must point to the same user as the JWT claims.
			// An alternative would be to use only the JWT claim to get the ID...
			user, err := s.GetUser(r.Context(), int32(id))
			if err != nil {
				http.Error(rw, "not found", http.StatusNotFound)
				return
			}

			ctx := context.WithValue(r.Context(), USERCONTEXT, user)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

// GET /user/{id}/holdings
func handleHoldingsRoute(svc *service.CustodianSvc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(USERCONTEXT).(*model.User)

		// we'll use a 30s timeout to fetch the data
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		custodians, err := svc.FetchFromCustodian(ctx, user.Custodians...)
		if err != nil {
			http.Error(rw, "custodian error", http.StatusInternalServerError)
			return
		}
		holdings := model.AggregateHoldings(custodians)

		rw.Header().Add("content-type", "application/json")
		encoder := json.NewEncoder(rw)
		encoder.Encode(holdings)
	}
}

// GET /user/{id}/custodian/{custId}/transactions?type=[0-3]&aggregate
func handleTransactionsRoute(svc *service.CustodianSvc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(USERCONTEXT).(*model.User)

		// we'll use a 30s timeout to fetch the data
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		scustId := chi.URLParam(r, "custId")
		custId, err := strconv.Atoi(scustId)
		if err != nil {
			http.Error(rw, "invalid custId in GET /user/{id}/custodian/{custId}/transactions?type=[0-3]", http.StatusBadRequest)
			return
		}

		// check this user has access to this custodian
		found := false
		for _, each := range user.Custodians {
			if each == int32(custId) {
				found = true
			}
		}
		if !found {
			http.Error(rw, "invalid custodian ID", http.StatusUnauthorized)
			return
		}

		custodians, err := svc.FetchFromCustodian(ctx, int32(custId))
		if err != nil {
			http.Error(rw, "custodian error", http.StatusInternalServerError)
			return
		}
		if len(custodians) != 1 {
			http.Error(rw, "custodian count error", http.StatusInternalServerError)
			return
		}
		custodian := custodians[0]

		var txl []*model.Transaction
		if t, hasType := r.URL.Query()["type"]; hasType {
			txtype, err := strconv.Atoi(t[0])
			if err != nil {
				http.Error(rw, "invalid tx type", http.StatusBadRequest)
				return
			}
			txl = custodian.FilterTransactionsByType(model.TransactionType(txtype))
		} else {
			txl = custodian.Transactions
		}

		rw.Header().Add("content-type", "application/json")
		encoder := json.NewEncoder(rw)

		// If we have the summary variable on the url, provide the asset list summary
		if _, hasSummary := r.URL.Query()["summary"]; hasSummary {
			al := model.NewAssetList()
			for _, tx := range txl {
				al.AddTransaction(tx)
			}
			encoder.Encode(al.GetAssets())
		} else { // otherwise provide the transactions only
			encoder.Encode(txl)
		}
	}
}

// GET /user/{id}
func handleUserRoute() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(USERCONTEXT).(*model.User)

		// we'll use a 30s timeout to fetch the data
		rw.Header().Add("content-type", "application/json")
		encoder := json.NewEncoder(rw)
		encoder.Encode(user)
	}
}

func init() {
	trackCmd.PersistentFlags().StringP("listen", "l", "0.0.0.0:9998", "the address to listen on")
	trackCmd.PersistentFlags().StringP("custodian", "c", "http://localhost:9999/custodian/", "the custodian service url")

	rootCmd.AddCommand(trackCmd)
}
