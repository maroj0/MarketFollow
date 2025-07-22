package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Stock struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Symbol       string             `json:"symbol" bson:"symbol"`
	Quantity     float64            `json:"quantity" bson:"quantity"`
	PurchaseDate time.Time          `json:"purchase_date" bson:"purchase_date"`
	BuyPrice    float64            `json:"buy_price" bson:"buy_price"`
	CurrentPrice float64            `json:"current_price,omitempty" bson:"-"`
}

type StockResponse struct {
	Symbol       string    `json:"symbol"`
	Quantity     float64   `json:"quantity"`
	BuyPrice    float64   `json:"buy_price"`
	CurrentPrice float64   `json:"current_price"`
	PurchaseDate time.Time `json:"purchase_date"`
	ProfitLoss  float64   `json:"profit_loss"`
	ProfitLossPct float64  `json:"profit_loss_percent"`
}

type PortfolioSummary struct {
	TotalInvested     float64         `json:"total_invested"`
	CurrentValue      float64         `json:"current_value"`
	TotalProfitLoss   float64         `json:"total_profit_loss"`
	TotalProfitLossPct float64        `json:"total_profit_loss_percent"`
	Stocks           []StockResponse `json:"stocks"`
}

var (
	mongoClient *mongo.Client
	collection  *mongo.Collection
)

func main() {
	// Cargar variables de entorno
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Configurar MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "your_mongodb_connection_string" {
		log.Fatal("Please set your MongoDB connection string in the .env file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	mongoClient = client
	collection = client.Database("stock_tracker").Collection("stocks")

	// Configurar el enrutador Gin
	router := gin.Default()

	// Configurar rutas
	api := router.Group("/api")
	{
		api.GET("/stocks", getStocks)
		api.POST("/stocks", addStock)
		api.GET("/stocks/summary", getPortfolioSummary)
	}

	// Configurar el puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configurar el host para que sea accesible desde la red
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	// Iniciar el servidor
	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Server running on %s\n", addr)
	log.Fatal(router.Run(addr))
}

// getStocks obtiene todas las acciones en la cartera
func getStocks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener las acciones"})
		return
	}
	defer cursor.Close(ctx)

	var stocks []Stock
	if err = cursor.All(ctx, &stocks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al decodificar las acciones"})
		return
	}

	c.JSON(http.StatusOK, stocks)
}

// addStock agrega una nueva acción a la cartera
type AddStockRequest struct {
	Symbol   string  `json:"symbol" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
}

func addStock(c *gin.Context) {
	var req AddStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obtener el precio actual de la acción
	currentPrice, err := getCurrentStockPrice(req.Symbol)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No se pudo obtener el precio para %s", req.Symbol)})
		return
	}

	stock := Stock{
		Symbol:       req.Symbol,
		Quantity:     req.Quantity,
		BuyPrice:    currentPrice,
		PurchaseDate: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, stock)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar la acción"})
		return
	}

	stock.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, stock)
}

// getPortfolioSummary obtiene un resumen de la cartera
func getPortfolioSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener las acciones"})
		return
	}
	defer cursor.Close(ctx)

	var stocks []Stock
	if err = cursor.All(ctx, &stocks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al decodificar las acciones"})
		return
	}

	if len(stocks) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No hay acciones en la cartera"})
		return
	}

	var summary PortfolioSummary
	var responses []StockResponse

	for _, stock := range stocks {
		currentPrice, err := getCurrentStockPrice(stock.Symbol)
		if err != nil {
			// Usar el precio de compra si no se puede obtener el actual
			currentPrice = stock.BuyPrice
		}

		invested := stock.Quantity * stock.BuyPrice
		currentValue := stock.Quantity * currentPrice
		profitLoss := currentValue - invested
		profitLossPct := 0.0
		if invested > 0 {
			profitLossPct = (profitLoss / invested) * 100
		}

		response := StockResponse{
			Symbol:        stock.Symbol,
			Quantity:      stock.Quantity,
			BuyPrice:     stock.BuyPrice,
			CurrentPrice:  currentPrice,
			PurchaseDate:  stock.PurchaseDate,
			ProfitLoss:    profitLoss,
			ProfitLossPct: profitLossPct,
		}

		summary.TotalInvested += invested
		summary.CurrentValue += currentValue
		responses = append(responses, response)
	}

	summary.TotalProfitLoss = summary.CurrentValue - summary.TotalInvested
	if summary.TotalInvested > 0 {
		summary.TotalProfitLossPct = (summary.TotalProfitLoss / summary.TotalInvested) * 100
	}
	summary.Stocks = responses

	c.JSON(http.StatusOK, summary)
}

// getCurrentStockPrice obtiene el precio actual de una acción usando Alpha Vantage
func getCurrentStockPrice(symbol string) (float64, error) {
	apiKey := os.Getenv("API_KEY_ALPHAVANTAGE")
	if apiKey == "" {
		return 0, fmt.Errorf("API key no configurada")
	}

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if quote, ok := result["Global Quote"]; ok {
		if priceStr, ok := quote["05. price"].(string); ok {
			return strconv.ParseFloat(priceStr, 64)
		}
	}

	return 0, fmt.Errorf("no se pudo obtener el precio")
}
