package controller

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"yy-event/configs"
	"yy-event/dtos"
	"yy-event/responses"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var memberCollection *mongo.Collection = configs.GetCollection(configs.DB, "member")
var validate = validator.New()

func Registrasi(c *fiber.Ctx) error {
	// berfungsi untuk memberikan timeout pada operasi yang mungkin akan memakan waktu, time out diberikan sebesar 10 detik
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// mengaktifkan fungsi cancel untuk operasi yang melebihi 10 detik
	defer cancel()

	var member dtos.Member
	// json request body akan di parse/dimasukkan ke struct member
	if err := c.BodyParser(&member); err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.MemberResponse{
			Status:  http.StatusBadRequest,
			Message: "Failed to parse the request body",
			Data:    &fiber.Map{"error": err.Error()},
		})
	}

	// struct member akan di validasi dengan memberValidate (cek kolom yang diset "required")
	if validationErr := validate.Struct(&member); validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.MemberResponse{
			Status:  http.StatusBadRequest,
			Message: "Validation failed",
			Data:    &fiber.Map{"error": validationErr.Error()},
		})
	}

	//membuat struct baru untuk dimasukkan ke database (bisa juga langsung menggunakan struct member yang sudah ada)
	newMember := dtos.Member{
		Nama:       member.Nama,
		Email:      member.Email,
		NoTelp:     member.NoTelp,
		Company:    member.Company,
		Region:     member.Region,
		Is_present: false,
	}

	//memasukkan struct newMember ke database
	_, err := memberCollection.InsertOne(ctx, newMember)
	if err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(responses.MemberResponse{
			Status:  http.StatusUnprocessableEntity,
			Message: "Failed to insert member",
			Data:    &fiber.Map{"error": err.Error()},
		})
	}

	endpoint := fmt.Sprintf("%s%s?sheet=%s", os.Getenv("BASE_URL"), os.Getenv("SHEETS_KEY"), os.Getenv("SHEETS_REGISTRATION"))
	body := []byte(fmt.Sprintf(`{"nama": "%s", "email": "%s", "region": "%s", "noTelp": "%s", "company": "%s"}`, member.Nama, member.Email, member.Region, member.NoTelp, member.Company))
	payload := bytes.NewReader(body)

	client := &http.Client{}

	request, err := http.NewRequest("POST", endpoint, payload)
	if err != nil {
		return c.Status(http.StatusInternalServerError).
			JSON(responses.MemberResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data:    &fiber.Map{"data": err.Error()}})
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return c.Status(http.StatusInternalServerError).
			JSON(responses.MemberResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data:    &fiber.Map{"data": err.Error()}})
	}
	defer response.Body.Close()

	//return successful response
	return c.Status(http.StatusCreated).JSON(responses.MemberResponse{
		Status:  http.StatusCreated,
		Message: "Member created successfully",
		Data:    &fiber.Map{"member email": newMember.Email},
	})
}

func Presensi(c *fiber.Ctx) error {
	// berfungsi untuk memberikan timeout pada operasi yang mungkin akan memakan waktu, time out diberikan sebesar 10 detik
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// mengaktifkan fungsi cancel untuk operasi yang melebihi 10 detik
	defer cancel()

	var presensi dtos.Presensi
	// json request body akan di parse/dimasukkan ke struct member
	if err := c.BodyParser(&presensi); err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.MemberResponse{
			Status:  http.StatusBadRequest,
			Message: "Failed to parse the request body",
			Data:    &fiber.Map{"error": err.Error()},
		})
	}

	// struct member akan di validasi dengan memberValidate (cek kolom yang diset "required")
	if validationErr := validate.Struct(&presensi); validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.MemberResponse{
			Status:  http.StatusBadRequest,
			Message: "Validation failed",
			Data:    &fiber.Map{"error": validationErr.Error()},
		})
	}

	var member dtos.Member
	err := memberCollection.FindOne(ctx, bson.M{"email": presensi.Email}).Decode(&member)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": http.StatusNotFound, "message": "Email is not exist in database"})
	}

	if member.Is_present == true {
		return c.Status(http.StatusInternalServerError).
			JSON(responses.MemberResponse{
				Status:  http.StatusUnprocessableEntity,
				Message: "Member is present",
				Data:    nil})
	}

	update := bson.M{"is_present": true}

	result, err := memberCollection.UpdateOne(ctx, bson.M{"email": member.Email}, bson.M{"$set": update})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.MemberResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	if result.MatchedCount == 1 && result.ModifiedCount == 1 {
		endpoint := fmt.Sprintf("%s%s?sheet=%s", os.Getenv("BASE_URL"), os.Getenv("SHEETS_KEY"), os.Getenv("SHEETS_ATTENDANCE"))
		body := []byte(fmt.Sprintf(`{"nama": "%s", "email": "%s", "region": "%s", "noTelp": "%s", "company": "%s"}`, member.Nama, member.Email, member.Region, member.NoTelp, member.Company))
		payload := bytes.NewReader(body)

		client := &http.Client{}

		request, err := http.NewRequest("POST", endpoint, payload)
		if err != nil {
			return c.Status(http.StatusInternalServerError).
				JSON(responses.MemberResponse{
					Status:  http.StatusInternalServerError,
					Message: "error",
					Data:    &fiber.Map{"data": err.Error()}})
		}

		request.Header.Set("Content-Type", "application/json")

		response, err := client.Do(request)
		if err != nil {
			return c.Status(http.StatusInternalServerError).
				JSON(responses.MemberResponse{
					Status:  http.StatusInternalServerError,
					Message: "error",
					Data:    &fiber.Map{"data": err.Error()}})
		}
		defer response.Body.Close()

		return c.Status(http.StatusOK).JSON(responses.MemberResponse{Status: http.StatusOK, Message: "success"})

	} else {
		return c.Status(http.StatusNotFound).JSON(responses.MemberResponse{Status: http.StatusNotFound, Message: "Document not found or not modified", Data: nil})
	}
}
