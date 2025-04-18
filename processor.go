package main

import (
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type idResponse struct {
	ID string `json:"idResponse"`
}

type pointsResponse struct {
	Points int64 `json:"pointsResponse"`
}

type item struct {
	Description string  `json:"shortDescription"`
	Price       float64 `json:"price,string"`
}

type receipt struct {
	Retailer string `json:"retailer"`
	Date     string `json:"purchaseDate"`
	Time     string `json:"purchaseTime"`
	Items    []item `json:"items"`
	Total    string `json:"total"`
}

// Receipts Cache / DB for the receipts
var Receipts = make(map[string]receipt)
var StartTime, _ = time.Parse("15:04", "14:00")
var EndTime, _ = time.Parse("15:04", "16:00")

func processReceipt(c *gin.Context) {
	var receipt receipt
	err := c.BindJSON(&receipt)

	if err != nil {
		c.JSON(http.StatusBadRequest, "The receipt is invalid")
		return
	}

	// Use os/sys command to generate UUID, need to trim trailing new line character
	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		c.JSON(http.StatusBadRequest, "The receipt is invalid")
		return
	}

	receiptId := string(uuid)
	receiptId = strings.Trim(receiptId, "\n")
	Receipts[receiptId] = receipt
	c.JSON(http.StatusOK, idResponse{ID: receiptId})

}

func getReceipt(c *gin.Context) {
	id := c.Param("idResponse")

	if receipt, found := Receipts[id]; found {
		points := 0
		points += scoreRetailer(receipt.Retailer)
		points += scoreReceiptTotal(receipt.Total)
		points += scoreItemCount(receipt.Items)
		points += scoreDay(receipt.Date)
		points += scoreTime(receipt.Time)
		points += scoreItemDescription(receipt)
		c.JSON(http.StatusOK, pointsResponse{Points: int64(points)})
	} else {
		c.JSON(http.StatusNotFound, "No receipt found for that id")
	}

}

// Add one point for every alphanumeric character in the retailer name
func scoreRetailer(retailer string) int {
	sum := 0
	for _, char := range retailer {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			sum++
		}
	}
	return sum
}

// Check if the total is a round dollar amount with no cents or if the total is a multiple of 0.25
func scoreReceiptTotal(total string) int {
	// Split the dollars from cents
	currencyTotal := strings.Split(total, ".")
	if len(currencyTotal) != 2 {
		return 0
	}

	cents, err := strconv.Atoi(currencyTotal[1])
	if err != nil {
		return 0
	}

	points := 0
	// 50 pointsResponse if the total is a round dollar amount with no cents
	if cents == 0 {
		points += 50
	}

	// 25 pointsResponse if the total is a multiple of 0.25
	if cents%25 == 0 {
		points += 25
	}

	return points
}

// Add points for every two items
func scoreItemCount(items []item) int {
	itemSum := 0
	for i := 0; i < len(items); i += 2 {
		itemSum += 5
	}
	return itemSum
}

// Add points if the day of the purchase is odd
func scoreDay(date string) int {
	parsedDate, _ := time.Parse("2006-01-02", date)
	if parsedDate.Day()%2 == 0 {
		return 0
	}
	return 6
}

// Add points if the purchase was between 2:00 PM and 4:00 PM
func scoreTime(timestamp string) int {
	parsedTime, _ := time.Parse("15:04", timestamp)
	if parsedTime.After(StartTime) && parsedTime.Before(EndTime) {
		return 10
	}
	return 0
}

// Add points if the trimmed length of the item description is a multiple of 3, multiply the price by 0.2
// and round up to the nearest integer. The result is the number of pointsResponse earned
func scoreItemDescription(receipt receipt) int {
	points := 0
	for _, item := range receipt.Items {
		trimmed := strings.Trim(item.Description, " ")

		if len(trimmed)%3 == 0 {
			priceMultiplied := 0.2 * item.Price
			points += int(math.Ceil(priceMultiplied))
		}
	}
	return points
}
