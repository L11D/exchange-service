package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html/charset"
)

// Валюта
type Valute struct {
	XMLName      xml.Name `xml:"Valute"`
	ID           string   `xml:"ID,attr"`
	NumCode      string   `xml:"NumCode"`
	CharCode     string   `xml:"CharCode"`
	Nominal      int      `xml:"Nominal"`
	Name         string   `xml:"Name"`
	ValueStr     string   `xml:"Value"`
	VunitRateStr string   `xml:"VunitRate"`
	Value        float64  `xml:"-"`
	VunitRate    float64  `xml:"-"`
}

// Корневой элемент
type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valutes []Valute `xml:"Valute"`
}

func getDataFromCBR() io.ReadCloser {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.cbr.ru/scripts/XML_daily.asp", nil)
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка выполнения запроса:", err)
		return nil
	}
	return resp.Body
}

func parseXML(body io.ReadCloser) ValCurs {
	decoder := xml.NewDecoder(body)
	decoder.CharsetReader = charset.NewReaderLabel

	var valCurs ValCurs
	err := decoder.Decode(&valCurs)
	if err != nil {
		fmt.Println("Ошибка разбора XML:", err)
		return ValCurs{}
	}

	for i, v := range valCurs.Valutes {
		valCurs.Valutes[i].Value = parseFloat(v.ValueStr)
		valCurs.Valutes[i].VunitRate = parseFloat(v.VunitRateStr)
	}
	return valCurs
}

func parseFloat(value string) float64 {
	value = strings.Replace(value, ",", ".", -1) // Заменяем запятую на точку
	f, err := strconv.ParseFloat(value, 32)
	if err != nil {
		fmt.Println("Ошибка конвертации:", err)
		return 0.0
	}
	return f
}

func convert(valuteFrom string, valuteTo string, amount float64) (float64, float64, error) {
	responseBody := getDataFromCBR()
	valutes := parseXML(responseBody).Valutes
	amountRubles, err := convertToRUB(valutes, valuteFrom, amount)
	if err != nil {
		return 0, 0, err
	}
	result, err := convertFromRUBToValute(valutes, valuteTo, amountRubles)
	if err != nil {
		return 0, 0, err
	}
	return result, amount/result, nil
}

func convertFromRUBToValute(valutes []Valute, valuteTo string, amount float64) (float64, error) {
	if valuteTo == "RUB" {
		return amount, nil
	}
	for _, valute := range valutes {
		if valute.CharCode == valuteTo {
			return amount / valute.VunitRate, nil
		}
	}
	return 0.0, errors.New("Valute not found")
}

func convertToRUB(valutes []Valute, valuteFrom string, amount float64) (float64, error) {
	if valuteFrom == "RUB" {
		return amount, nil
	}
	for _, valute := range valutes {
		if valute.CharCode == valuteFrom {
			return amount * valute.VunitRate, nil
		}
	}
	return 0.0, errors.New("Valute not found")
}

type ExchangeRequest struct {
	CurrencyFrom string  `json:"currencyFrom" binding:"required"`
	CurrencyTo   string  `json:"currencyTo" binding:"required"`
	Amount       float64 `json:"amount" binding:"required"`
}

func getExchange(c *gin.Context) {
	var request ExchangeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, rate, err := convert(request.CurrencyFrom, request.CurrencyTo, request.Amount)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"amount": result, "rate": rate})
}

func getCurrency(c *gin.Context) {
	responseBody := getDataFromCBR()
	valutes := parseXML(responseBody).Valutes
	var valutesNames []string
	valutesNames = append(valutesNames, "RUB")
	for _, valute := range valutes {
		valutesNames = append(valutesNames, valute.CharCode)
	}
	sort.Strings(valutesNames)
	c.JSON(http.StatusOK, valutesNames)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router := gin.Default()
	router.POST("/exchange", getExchange)
	router.GET("/currencies", getCurrency)
	router.Run(":" + port)
	fmt.Println("server is running on port", port)
}
