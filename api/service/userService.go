package service

import (
	"chat-app/pkg/helper"
	"chat-app/pkg/model"
	"chat-app/pkg/mongodb"
	"context"
	"fmt"
	"log"
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
		return err
	}

	if err := validate.Struct(user); err != nil {
		return err
	}

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		return fmt.Errorf("error occurred while checking for the email: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("this email or phone number already exists")
	}

	password := HashPassword(*user.Password)
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

	token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)

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

	// recordPerPage, err := strconv.Atoi(c.QueryParam("recordPerPage"))
	// if err != nil || recordPerPage < 1 {
	// 	recordPerPage = 10
	// }

	// page, err := strconv.Atoi(c.QueryParam("page"))
	// if err != nil || page < 1 {
	// 	page = 1
	// }

	// startIndex := (page - 1) * recordPerPage

	// matchStage := bson.D{{"$match", bson.D{{}}}}
	// projectStage := bson.D{
	// 	{"$project", bson.D{
	// 		{"_id", 0},
	// 		{"total_count", 1},
	// 		{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
	// 	}},
	// }

	// result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
	// 	matchStage, projectStage,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("error occurred while listing user items: %w", err)
	// }

	// var allUsers []bson.M
	// if err = result.All(ctx, &allUsers); err != nil {
	// 	return nil, fmt.Errorf("error occurred while decoding user items: %w", err)
	// }
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

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	if err != nil {
		return false, "login or password is incorrect"
	}
	return true, ""
}
