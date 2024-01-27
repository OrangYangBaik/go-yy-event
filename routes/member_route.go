package routes

import (
	controller "yy-event/controllers"

	"github.com/gofiber/fiber/v2"
)

func LoginRegis(app *fiber.App) {
	app.Post("/registrasi", controller.Registrasi)
	app.Post("/presensi", controller.Presensi)
}
