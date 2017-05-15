package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("bf", "BitFlyer CLI")
	app.Command("status", "Print account status", func(c *cli.Cmd) {
		c.Action = PrintTotalAssets
	})
	app.Command("history", "Print history for deposite and trade", func(c *cli.Cmd) {
		c.Action = PrintHistory
	})
	app.Run(os.Args)
}

func PrintTotalAssets() {
	api := New()

	bl, err := api.Getbalance()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	jpy, err := bl.Asset("JPY")
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	btc, err := bl.Asset("BTC")
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	tk, err := api.Ticker()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	b, _ := json.Marshal(struct {
		TotalAssets float64 `json:"total_assets"`
	}{
		TotalAssets: jpy.Amount + float64(tk.BestAsk)*btc.Amount,
	})
	fmt.Printf("%v\n", string(b))
}

func PrintHistory() {
	api := New()
	depo, err := api.Getdeposits()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	ins, err := api.Getcoinins()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	exc, err := api.Getexecutions()
	if err != nil {
	}

	b, _ := json.Marshal(struct {
		Deposits   []Deposit   `json:"deposits"`
		Coinins    []Coinin    `json:"coinins"`
		Executions []Execution `json:"executions"`
	}{
		Deposits:   depo.Deposits,
		Coinins:    ins.Coinins,
		Executions: exc.Executions,
	})
	fmt.Printf("%v\n", string(b))
}

func New() *Bitflyer {
	key, b := os.LookupEnv("BITFLYER_API_KEY")
	if key == "" || !b {
		log.Fatalf("BITFLYER_API_KEY is undefined")
	}
	secret, b := os.LookupEnv("BITFLYER_API_SECRET")
	if secret == "" || !b {
		log.Fatalf("BITFLYER_SECRET is undefined")
	}
	return &Bitflyer{
		key:      key,
		secret:   secret,
		endpoint: "https://api.bitflyer.jp",
	}
}

type Bitflyer struct {
	key      string
	secret   string
	endpoint string
}

func (b *Bitflyer) NewRequest(method, path string) (*http.Request, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	text := ts + method + path

	hash := hmac.New(sha256.New, []byte(b.secret))
	hash.Write([]byte(text))
	sign := hex.EncodeToString(hash.Sum(nil))

	req, err := http.NewRequest(method, b.endpoint+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json; charset=UTF-8")
	req.Header.Set("ACCESS-KEY", b.key)
	req.Header.Set("ACCESS-TIMESTAMP", ts)
	req.Header.Set("ACCESS-SIGN", sign)

	return req, nil
}

func SendRequest(req *http.Request, r interface{}) error {
	client := new(http.Client)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return err
	}

	e, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(e, r)
	return err
}

type Balance struct {
	Assets []Asset
}

func (b *Balance) Asset(code string /* ex) JPY, USD */) (*Asset, error) {
	for _, a := range b.Assets {
		if strings.EqualFold(a.CurrencyCode, code) {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("%v is missing", code)
}

type Asset struct {
	CurrencyCode string  `json:"currency_code"`
	Amount       float64 `json:"amount"`
	Available    float64 `json:"available"`
}

func (b *Bitflyer) Getbalance() (*Balance, error) {
	req, err := b.NewRequest("GET", "/v1/me/getbalance")
	if err != nil {
		return nil, err
	}

	r := &Balance{}
	return r, SendRequest(req, &r.Assets)
}

type Ticker struct {
	ProductCode     string  `json:"product_code"`
	Timestamp       string  `json:"timestamp"`
	TickID          float64 `json:"tick_id"`
	BestBid         float64 `json:"best_bid"`
	BestAsk         float64 `json:"best_ask"`
	BestBidSize     float64 `json:"best_bid_size"`
	BestAskSize     float64 `json:"best_ask_size"`
	TotalBidDepth   float64 `json:"total_bid_depth"`
	TotalAskDepth   float64 `json:"total_ask_depth"`
	Ltp             float64 `json:"ltp"`
	Volume          float64 `json:"volume"`
	VolumeByProduct float64 `json:"volume_by_product"`
}

func (b *Bitflyer) Ticker() (*Ticker, error) {
	req, err := b.NewRequest("GET", "/v1/ticker")
	if err != nil {
		return nil, err
	}

	r := &Ticker{}
	return r, SendRequest(req, &r)
}

type Deposits struct {
	Deposits []Deposit
}

type Deposit struct {
	ID           int     `json:"id"`
	OrderID      string  `json:"order_id"`
	CurrencyCode string  `json:"currency_code"`
	Amount       float64 `json:"amount"`
	Status       string  `json:"status"`
	EventDate    string  `json:"event_date"`
}

func (b *Bitflyer) Getdeposits() (*Deposits, error) {
	req, err := b.NewRequest("GET", "/v1/me/getdeposits")
	if err != nil {
		return nil, err
	}

	r := &Deposits{}
	return r, SendRequest(req, &r.Deposits)
}

type Coinin struct {
	ID           int     `json:"id"`
	OrderID      string  `json:"order_id"`
	CurrencyCode string  `json:"currency_code"`
	Amount       float64 `json:"amount"`
	Address      string  `json:"address"`
	TxHash       string  `json:"tx_hash"`
	Status       string  `json:"status"`
	EventDate    string  `json:"event_date"`
}

type Coinins struct {
	Coinins []Coinin
}

func (b *Bitflyer) Getcoinins() (*Coinins, error) {
	req, err := b.NewRequest("GET", "/v1/me/getcoinins")
	if err != nil {
		return nil, err
	}

	r := &Coinins{}
	return r, SendRequest(req, &r.Coinins)
}

type Execution struct {
	ID                     int     `json:"id"`
	ChildOrderID           string  `json:"child_order_id"`
	Side                   string  `json:"side"`
	Price                  int     `json:"price"`
	Size                   float64 `json:"size"`
	Commission             int     `json:"commission"`
	ExecDate               string  `json:"exec_date"`
	ChildOrderAcceptanceID string  `json:"child_order_acceptance_id"`
}

type Executions struct {
	Executions []Execution
}

func (b *Bitflyer) Getexecutions() (*Executions, error) {
	req, err := b.NewRequest("GET", "/v1/me/getexecutions")
	if err != nil {
		return nil, err
	}

	r := &Executions{}
	return r, SendRequest(req, &r.Executions)
}
