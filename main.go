package main

import (
	"log"
	"os"
	"yy-event/configs"
	"yy-event/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	// mengambil env variable PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "6000"
	}

	// Enable CORS for all origins with various options
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	//buat konfigurasi koneksi ke database
	configs.ConnectDB()

	//router
	routes.LoginRegis(app)

	// akses server side melalui IP dan port yang ditentukan (pada railway)
	log.Fatal(app.Listen("0.0.0.0:" + port))
}
