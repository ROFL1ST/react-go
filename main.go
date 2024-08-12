package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ROFL1ST/react-go/model"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func main() {
	fmt.Println("Hello Worlds")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file : ", err)
	}

	MONGO_URI := os.Getenv("MONGO_URI")
	ClientOption := options.Client().ApplyURI(MONGO_URI)
	client, err := mongo.Connect(context.Background(), ClientOption)

	if err != nil {
		log.Fatal(err)

	}
	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)

	}

	fmt.Println("Connected to mongodb")

	collection = client.Database("golang_db").Collection("todos")

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodos)
	app.Patch("/api/todos/:id", updateTodos)
	app.Delete("/api/todos/:id", deleteTodos)

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTodos(c *fiber.Ctx) error {
	var todos []model.Todo

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching todos")
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var todo model.Todo

		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}
	return c.JSON(todos)
}

func createTodos(c *fiber.Ctx) error {
	todo := new(model.Todo)

	if err := c.BodyParser(todo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "Please fill it correctly"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return nil
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(fiber.Map{"status": "Success", "data": todo})
}

func updateTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "Invalid id"})
	}

	filter := bson.M{"_id": ObjectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"Status": "Success", "message": "Todo's has been updated"})
}

func deleteTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	ObjectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Message": "Invalid id"})
	}

	filter := bson.M{"_id": ObjectID}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"Status": "Success", "message": "Todo's has been deleted"})

}
