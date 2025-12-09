package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"

	jwtware "github.com/gofiber/jwt/v2"
	"github.com/golang-jwt/jwt/v4"

	"database/sql"

	_ "github.com/lib/pq"
)

const (
	host = "localhost"
	port = 6543
	databaseName = "mydatabase"
	username = "myuser"
	password = "mypassword"
)

var db *sql.DB

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("load .env error")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
    "password=%s dbname=%s sslmode=disable",
    host, port, username, password, databaseName)

	sdb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	db = sdb

	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database: ", err)
	}

	fmt.Println("Successfully connected to the database!")
	
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	books = append(books, Book{ID: 1, Title: "1984", Author: "George Orwell"})
	books = append(books, Book{ID: 2, Title: "To Kill a Mockingbird", Author: "Harper Lee"})

	app.Post("/login", login)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}))

	app.Use(checkMiddleware)

	app.Get("/books", getBooks)
	app.Get("/books/:id", getBook)
	app.Post("/books", createBook)
	app.Put("/books/:id", updateBook)
	app.Delete("/books/:id", deleteBook)

	app.Post("/upload", uploadFile)
	app.Get("test-html", testHTML)

	app.Get("/config", getENV)

	app.Listen(":8080")
}

func uploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("File upload error")
	}
	err = c.SaveFile(file, "./uploads/"+file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not save file")
	}
	return c.SendString("File uploaded successfully: " + file.Filename)
}

func checkMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	if claims["role"] != "admin" {
		return fiber.ErrUnauthorized
	}
	return c.Next()
}	

func testHTML(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "Test HTML Rendering",
	})
}

func getENV(c *fiber.Ctx) error {
	if value, exist := os.LookupEnv("SECRET"); exist {
		return c.JSON(fiber.Map{
			"SECRET": value,
		})
	}
	return c.JSON(fiber.Map{
		"SECRET": "defaultConfig",
	})
}

type User struct {
	Email   string `json:"email"`
	Password string `json:"password"`
}
var memberUser = User{
	Email:    "test@example.com",
	Password: "1234",
}

func login(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if user.Email == "" || user.Password == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email and Password are required")
	}

	if memberUser.Email != user.Email || memberUser.Password != user.Password {
		return fiber.ErrUnauthorized
	}

	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
    claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["role"] = "admin" // example role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	
	return c.JSON(fiber.Map{
		"status": true,
		"message": "Login Success",
		"token": t,
	})
}

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func createProduct (product *Product) error {
	_, err := db.Exec("INSERT INTO products (name, price) VALUES ($1, $2)", product.Name, product.Price)
	if err != nil {
		log.Fatal("Error inserting product: ", err)
	}

	return err
}