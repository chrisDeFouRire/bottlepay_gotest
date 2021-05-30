package cmd

import (
	"net/http"
	"testing"

	"github.com/bottlepay/portfolio-data/model"
	"github.com/gavv/httpexpect/v2"
)

func Test_prerequisites(t *testing.T) {
	e := httpexpect.New(t, "http://localhost:9999/")

	// this should be enough to ensure the tests run against the correct data
	e.GET("/custodian/1").
		Expect().JSON().Object().Value("assets").Array().First().Object().Value("balance").
		Equal("31.12874002")
	e.GET("/custodian/2").
		Expect().JSON().Object().Value("assets").Array().First().Object().Value("balance").
		Equal("40.92631191")
	e.GET("/custodian/3").
		Expect().JSON().Object().Value("assets").Array().First().Object().Value("balance").
		Equal("13.40737117")
	e.GET("/custodian/4").
		Expect().JSON().Object().Value("assets").Array().First().Object().Value("balance").
		Equal("9.10921833")
}

func Test_handleUserRoute(t *testing.T) {
	e := httpexpect.New(t, "http://localhost:9998/")

	user := e.GET("/user/1").
		Expect().
		Status(http.StatusOK).JSON().Object()

	user.Value("id").Number().Equal(1)
	user.Value("custodians").Array().Length().Equal(4)

	e.GET("/user/2").
		Expect().
		Status(http.StatusNotFound)

	e.GET("/user/alpha").
		Expect().
		Status(http.StatusNotFound)

}

func Test_handleHoldingsRoute(t *testing.T) {
	e := httpexpect.New(t, "http://localhost:9998/")

	holdings := e.GET("/user/1/holdings").
		Expect().
		Status(http.StatusOK).JSON().Array()

	holdings.Length().Equal(2)
	holdings.First().Object().Keys().ContainsOnly("balance", "code")
	holdings.Last().Object().Keys().ContainsOnly("balance", "code")

	e.GET("/user/2/holdings").
		Expect().
		Status(http.StatusNotFound)

}

// This integration test really requires to run against the generator with --time=0 and
// the provided data/state.json file
func Test_handleTransactionsRoute(t *testing.T) {
	e := httpexpect.New(t, "http://localhost:9998/")

	e.GET("/user/1/custodian/999/transactions").
		Expect().
		Status(http.StatusUnauthorized)

	e.GET("/user/3/custodian/999/transactions").
		Expect().
		Status(http.StatusNotFound)

	txl := e.GET("/user/1/custodian/1/transactions").
		Expect().Status(http.StatusOK).JSON().Array()
	txl.Length().Equal(37)

	txl = e.GET("/user/1/custodian/2/transactions").
		Expect().Status(http.StatusOK).JSON().Array()
	txl.Length().Equal(50)

	txl = e.GET("/user/1/custodian/1/transactions").WithQuery("type", model.ExternalDeposit).
		Expect().Status(http.StatusOK).JSON().Array()
	txl.Length().Equal(9)
	txl = e.GET("/user/1/custodian/1/transactions").WithQuery("type", model.ForeignTransfer).
		Expect().Status(http.StatusOK).JSON().Array()
	txl.Length().Equal(24)
	txl = e.GET("/user/1/custodian/2/transactions").WithQuery("type", model.InternalAssetExchange).
		Expect().Status(http.StatusOK).JSON().Array()
	txl.Length().Equal(12)

	txl = e.GET("/user/1/custodian/2/transactions").WithQuery("type", model.ExternalWithdrawal).
		Expect().Status(http.StatusOK).JSON().Array()
	txl.Length().Equal(6) // hand checked

	assets := e.GET("/user/1/custodian/2/transactions").
		WithQuery("type", model.ExternalWithdrawal).
		WithQuery("summary", true).
		Expect().Status(http.StatusOK).JSON().Array()
	assets.Length().Equal(2)
	assets.First().Object().Value("code").Equal("BTC")
	assets.First().Object().Value("balance").Equal("7.6262799625") // hand checked
	assets.Last().Object().Value("code").Equal("GBP")
	assets.Last().Object().Value("balance").Equal("43046.9044724478") // hand checked
}
