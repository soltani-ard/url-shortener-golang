package routes

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/soltani-ard/url-shortener-golang/database"
	"github.com/soltani-ard/url-shortener-golang/helpers"
	"os"
	"strconv"
	"time"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset int           `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {

	// parse json
	body := new(request)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{
				"error": "cannot parse JSON.",
			})
	}

	// limit rate
	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.IP()).Result()
	fmt.Println(err)
	if err == redis.Nil { // record not found in database => set new value
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else { // record exist
		val, _ = r2.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 { // limit
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).
				JSON(fiber.Map{
					"error":       "rate limit.",
					"limit_reset": limit / time.Nanosecond / time.Minute,
				})
		}
	}
	// check url is valid
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{
				"error": "Invalid URL.",
			})
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(fiber.Map{
				"error": "domain error.",
			})
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	var id string
	// check user send data or not
	if body.CustomShort == "" { // create new data
		id = uuid.New().String()[:6]
	} else { // use user url
		id = body.CustomShort
	}
	// connect to database
	r := database.CreateClient(0)
	defer r.Close()
	// check id in database
	val, _ = r.Get(database.Ctx, id).Result()
	// use of other users
	if val != "" {
		return c.Status(fiber.StatusForbidden).
			JSON(fiber.Map{"error": "url custom short is already in use."})
	}

	// check expiry
	if body.Expiry == 0 {
		body.Expiry = 24
	}

	// set new data
	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "unable to connect to server."})
	}

	// create response and edit default value
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	// decrement use service for client
	r2.Decr(database.Ctx, c.IP())

	// update remaining request number for call api per user
	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	// update ttl
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = int(ttl / time.Nanosecond / time.Minute)

	// update custom short
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
