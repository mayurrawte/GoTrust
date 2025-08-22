package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mayurrawte/gotrust"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User document structure in MongoDB
type mongoUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	Name      string             `bson:"name"`
	AvatarURL string             `bson:"avatar_url,omitempty"`
	Provider  string             `bson:"provider"`
	Password  string             `bson:"password,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

// MongoUserStore implements gotrust.UserStore
type MongoUserStore struct {
	collection *mongo.Collection
}

func NewMongoUserStore(db *mongo.Database) (*MongoUserStore, error) {
	collection := db.Collection("users")

	// Create unique index on email
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		return nil, fmt.Errorf("failed to create email index: %w", err)
	}

	return &MongoUserStore{collection: collection}, nil
}

func (s *MongoUserStore) CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error {
	doc := mongoUser{
		ID:        primitive.NewObjectID(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Provider:  user.Provider,
		Password:  hashedPassword,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	result, err := s.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (s *MongoUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
	var doc mongoUser
	
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, "", fmt.Errorf("user not found")
		}
		return nil, "", err
	}

	user := &gotrust.User{
		ID:        doc.ID.Hex(),
		Email:     doc.Email,
		Name:      doc.Name,
		AvatarURL: doc.AvatarURL,
		Provider:  doc.Provider,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}

	return user, doc.Password, nil
}

func (s *MongoUserStore) GetUserByID(ctx context.Context, userID string) (*gotrust.User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	var doc mongoUser
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &gotrust.User{
		ID:        doc.ID.Hex(),
		Email:     doc.Email,
		Name:      doc.Name,
		AvatarURL: doc.AvatarURL,
		Provider:  doc.Provider,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}, nil
}

func (s *MongoUserStore) UpdateUser(ctx context.Context, user *gotrust.User) error {
	objectID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
			"updated_at": time.Now(),
		},
	}

	result, err := s.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (s *MongoUserStore) UserExists(ctx context.Context, email string) (bool, error) {
	count, err := s.collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func main() {
	// MongoDB connection
	mongoURI := "mongodb://localhost:27017"
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(context.Background())

	// Ping to verify connection
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	// Setup database and user store
	db := client.Database("gotrust_example")
	userStore, err := NewMongoUserStore(db)
	if err != nil {
		log.Fatal("Failed to create user store:", err)
	}

	// GoTrust setup
	config := gotrust.NewConfig()
	if config.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	sessionStore := gotrust.NewMemorySessionStore()
	authService := gotrust.NewAuthService(config, userStore, sessionStore)

	// Echo setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register auth routes
	handlers := gotrust.NewAuthHandlers(authService, config)
	handlers.RegisterRoutes(e, "/auth")

	// Home route
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "GoTrust MongoDB Example",
			"database": "MongoDB",
		})
	})

	// Protected routes
	api := e.Group("/api")
	api.Use(authService.AuthMiddleware())

	api.GET("/profile", func(c echo.Context) error {
		userID, _ := gotrust.GetUserFromContext(c)
		return c.JSON(200, map[string]string{
			"user_id": userID,
			"message": "Protected profile endpoint",
		})
	})

	// Start server
	port := ":8080"
	log.Printf("Server starting on http://localhost%s (MongoDB: %s)", port, mongoURI)
	if err := e.Start(port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}