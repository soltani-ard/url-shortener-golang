package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/soltani-ard/url-shortener-golang/database"
)

func ResolveURL(c *fiber.Ctx) error {
	// get url
	url := c.Params("url")

	// get of database
	r := database.CreateClient(0)
	// close database ofter execute function
	defer r.Close()
	// get value of redis
	value, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil { // not found
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"error": "short not found in database."})
	} else if err != nil { // internal error
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "cannot connect to database."})
	}

	// create counter per request for limit
	rInr := database.CreateClient(1)
	defer rInr.Close()
	_ = rInr.Incr(database.Ctx, "counter")
	return c.Redirect(value, 301)
}
