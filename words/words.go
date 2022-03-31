package words

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	log "encore.dev/rlog"
	"encore.dev/storage/sqldb"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// encore:api public path=/word
func TodaysWord(ctx context.Context) (*Response, error) {
	startDate := time.Date(2022, time.March, 26, 4, 0, 0, 0, time.UTC)
	today := time.Now().In(time.UTC)

	days := int(today.Sub(startDate).Hours() / 24)
	if days >= len(words) {
		log.Error("Words list out of indices")
		return nil, errors.New("Words list out of indices")
	}

	log.Info(fmt.Sprintf("Today's word is: %s", words[days]))
	return &Response{words[days]}, nil
}

// encore:api private path=/send-messages
func SendMessages(ctx context.Context) error {
	phoneNumbers, err := GetPhoneNumbers(ctx)
	if err != nil {
		return err
	}

	todaysWord, err := TodaysWord(ctx)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("%s 🖕", strings.ToUpper(todaysWord.Message))
	for _, to := range phoneNumbers.Numbers {
		if err := sendMessage(*to, msg); err != nil {
			log.Error(err.Error(),
				"to", to)
		} else {
			log.Info("SMS sent successfully!",
				"to", to)
		}
	}

	return nil
}

func sendMessage(to, msg string) error {
	// Set env variables for use by client
	os.Setenv("TWILIO_ACCOUNT_SID", secrets.TwilioAccountSID)
	os.Setenv("TWILIO_AUTH_TOKEN", secrets.TwilioAuthToken)

	client := twilio.NewRestClient()

	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(secrets.TwilioPhoneNumber)
	params.SetBody(msg)

	_, err := client.ApiV2010.CreateMessage(params)
	return err
}

// encore:api private path=/add-number/:number
func AddPhoneNumber(ctx context.Context, number string) error {
	_, err := sqldb.Exec(ctx, `
        INSERT INTO phone_numbers (phone_number)
        VALUES ($1)
    `, number)

	if err != nil && strings.Contains(err.Error(), "duplicate key value") {
		log.Info("Number already in database, not adding.",
			"number", number)
		return nil
	} else if err != nil {
		log.Error("Number couldn't be added to database.",
			"number", number,
			"error", err.Error())
		return err
	} else {
		log.Info("Number added to database.",
			"number", number)
		return nil
	}
}

// encore:api private path=/get-numbers
func GetPhoneNumbers(ctx context.Context) (*Numbers, error) {
	rows, err := sqldb.Query(ctx, `
        SELECT phone_number FROM phone_numbers
    `)
	if err != nil {
		log.Error("Failed to fetch numbers from database.",
			"error", err.Error())
		return nil, err
	}

	numbers := []*string{}
	for rows.Next() {
		var num string
		err := rows.Scan(&num)
		if err != nil {
			log.Error("Error scanning result from database",
				"error", err.Error())
			return nil, err
		}
		numbers = append(numbers, &num)
	}

	return &Numbers{numbers}, nil
}

// encore:api private path=/remove-number/:number
func RemovePhoneNumber(ctx context.Context, number string) error {
	_, err := sqldb.Exec(ctx, `
        DELETE FROM phone_numbers
        WHERE phone_number = ($1)
    `, number)

	if err != nil {
		log.Error("Error deleting number from databse.",
			"number", number,
			"error", err.Error())
		return err
	} else {
		log.Info("Deleted number from database",
			"number", number)
		return nil
	}
}

type Numbers struct {
	Numbers []*string
}

type Response struct {
	Message string
}

var secrets struct {
	TwilioPhoneNumber string
	TwilioAccountSID  string
	TwilioAuthToken   string
}
