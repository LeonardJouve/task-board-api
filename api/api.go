package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func getParamInt(c *fiber.Ctx, name string) (int, bool) {
	param, err := c.ParamsInt(name)
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": fmt.Sprintf("invalid %s", name),
		})
		return 0, false
	}

	return param, true
}

func getQueryUIntArray(c *fiber.Ctx, name string) ([]uint, bool) {
	array := []uint{}
	for _, value := range strings.Split(c.Query(name), ",") {
		number, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("invalid %s", name),
			})
			return []uint{}, false
		}
		array = append(array, uint(number))
	}

	return array, true
}
