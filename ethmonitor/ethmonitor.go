package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

type result struct {
	Data struct {
		Base     string `json:"base"`
		Currency string `json:"currency"`
		Amount   string `json:"amount"`
	} `json:"data"`
	Warnings []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
	} `json:"warnings"`
}

func getEthPrice() float64 {
	url := "https://api.coinbase.com/v2/prices/eth-usd/spot?quote=true"

	r, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Body.Close()

	data := result{}
	json.NewDecoder(r.Body).Decode(&data)

	amount, err := strconv.ParseFloat(data.Data.Amount, 64)
	if err != nil {
		log.Fatalln(err)
	}

	return amount
}

func getClient(ctx context.Context) (*http.Client, error) {
	data, err := ioutil.ReadFile("credential.json")
	if err != nil {
		log.Fatal(err)
	}

	conf, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(oauth2.NoContext)
	return client, nil
}

func update() {
	ctx := context.Background()

	c, err := getClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sheetsService, err := sheets.New(c)
	if err != nil {
		log.Fatal(err)
	}

	// The ID of the spreadsheet to update.
	spreadsheetId := "1CJM-GAnXG5YTFvDOybJ0-FOHH1ZHT9SHkyNh8uAW0aQ"

	// The A1 notation of the values to update.
	range2 := "K5"

	newPrice := getEthPrice()

	rb := &sheets.ValueRange{
		// will be replaced.
		Values: [][]interface{}{
			{newPrice},
		},
	}

	valueService := sheetsService.Spreadsheets.Values
	updateCall := valueService.Update(spreadsheetId, range2, rb)
	updateCall = updateCall.Context(ctx)
	updateCall = updateCall.ValueInputOption("USER_ENTERED")

	for {
		_, err = updateCall.Do()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Changed to ", newPrice)

		time.Sleep(time.Second)

		newPrice = getEthPrice()
		rb.Values = [][]interface{}{
			{newPrice},
		}
	}
}

func main() {
	update()
}
