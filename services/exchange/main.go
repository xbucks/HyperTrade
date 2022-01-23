package main

import (
	"os"

	"exchange/db"
	"exchange/internal"

	"github.com/adshao/go-binance/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

var interval string = "1m"

var symbol = "ETHUSDT"

func main() {
	DB := db.New()

	wait := make(chan bool)
	env := internal.GetEnv()

	pubsub := internal.NewPubSub(env.NatsUrl, env.NatsUser, env.NatsPass)
	defer pubsub.Close()

	bex := internal.NewBinance(env.BinanceApiKey, env.BinanceApiSecretKey, pubsub, env.BinanceTestnet)
	bex.PrintAccountInfo()

	go bex.Kline(symbol, interval)

	pubsub.Subscribe(internal.DataFrameEvent, func(p internal.DataFrameEventPayload) {
		ListenTrade(DB, p.Kline, p.Signal)
	})

	<-wait
}

// TODO: refactor this
func ListenTrade(DB db.DB, kline internal.Kline, signal internal.Signal) {
	side := getSide(signal)

	if side == "" {
		return
	}

	symbol := kline.Symbol

	log.Trace().Str("symbol", symbol).Interface("side", side).Msg("Trade.Listen")

	position := DB.GetPosition(symbol)
	var holding bool = position.Symbol != ""

	config := DB.GetConfig(symbol)

	allowedAmt := config.AllowedAmount
	closePrice := kline.Close
	quantity := internal.GetMinQuantity(allowedAmt, closePrice)

	switch side {
	case binance.SideTypeBuy:
		if holding {
			log.Warn().Bool("holding", holding).Msg("Trade.Buy.Skip")
			return
		}

		// TODO: Execute Buy
		DB.CreatePosition(symbol, closePrice, quantity)
		log.Trace().Float64("price", closePrice).Float64("quantity", quantity).Msg("Trade.Buy.Complete")

	case binance.SideTypeSell:
		if !holding {
			log.Warn().Bool("holding", holding).Msg("Trade.Sell.Skip")
			return
		}

		// TODO: Execute Sell
		entry := position.Price
		DB.DeletePosition(symbol)
		DB.CreateTrade(symbol, entry, closePrice, quantity)
		log.Trace().Float64("price", closePrice).Float64("quantity", quantity).Msg("Trade.Sell.Complete")
	default:
	}
}

func getSide(signal internal.Signal) binance.SideType {
	var side binance.SideType

	if signal.Buy && !signal.Sell {
		side = binance.SideTypeBuy
	} else if signal.Sell && !signal.Buy {
		side = binance.SideTypeSell
	}

	return side
}

func isEmpty(value interface{}) bool {
	return value == ""
}
