package service

import (
	"chat-app/pkg/helper"
	"chat-app/pkg/model"
	"chat-app/pkg/mongodb"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = mongodb.OpenCollection("user")
var validate = validator.New()

func SignUp(c echo.Context) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user model.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": fmt.Sprintf("Request parsing failed: %s", err.Error())})
	}

	if err := validate.Struct(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": fmt.Sprintf("Request validation failed: %s", err.Error())})
	}

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		return fmt.Errorf("error occurred while checking for the email: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("this email already exists")
	}

	password, err := HashPassword(*user.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	user.Password = &password

	user.Created_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to parse created_at time: %w", err)
	}
	user.Updated_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to parse updated_at time: %w", err)
	}
	user.ID = primitive.NewObjectID()
	user.User_id = user.ID.Hex()

	token, refreshToken, err := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
	if err != nil {
		return fmt.Errorf("failed to generate tokens: %w", err)
	}
	user.Token = &token
	user.Refresh_Token = &refreshToken

	_, insertErr := userCollection.InsertOne(ctx, user)
	if insertErr != nil {
		return fmt.Errorf("user item not created: %w", insertErr)
	}

	return nil
}

func GetUser(c echo.Context) (model.User, error) {
	userId := c.Param("user_id")

	if err := helper.MatchUserTypeTOUId(c, userId); err != nil {
		return model.User{}, err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user model.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func Login(c echo.Context) (model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user model.User
	if err := c.Bind(&user); err != nil {
		return model.User{}, err
	}

	var foundUser model.User
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		return model.User{}, fmt.Errorf("user not found, login seems to be incorrect: %w", err)
	}

	passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
	if !passwordIsValid {
		return model.User{}, fmt.Errorf(msg)
	}

	token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
	helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

	return foundUser, nil
}

func GetUsers(c echo.Context) ([]model.User, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var documents []model.User
	cursor, err := userCollection.Find(ctx, bson.M{})
	if err != nil {
		return documents, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &documents); err != nil {
		return documents, err
	}

	return documents, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	if err != nil {
		return false, "login or password is incorrect"
	}
	return true, ""
}
