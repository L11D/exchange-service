package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	// "github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

// type album struct {
// 	ID     string  `json:"idddd"`
// 	Title  string  `json:"title"`
// 	Artist string  `json:"artist"`
// 	Price  float64 `json:"price"`
// }

// var albums = []album{
// 	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
// 	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
// 	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
// }

// func getAlbums(c *gin.Context) {
// 	c.IndentedJSON(http.StatusOK, albums)
// }

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

func convert(valuteFrom string, valuteTo string, amount float64) float64 {
	responseBody := getDataFromCBR()
	valutes := parseXML(responseBody).Valutes
	amountRubles, err := convertToRUB(valutes, valuteFrom, amount)
	if err != nil {
		
	}
	result, err := convertFromRUBToValute(valutes, valuteTo, amountRubles)
	if err != nil {
		
	}
	return result
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

func main() {
	valuteFrom := "USD"
	valuteTo := "EUR1"
	var amount float64 = 1000
	result := convert(valuteFrom, valuteTo, amount)

	fmt.Printf("%.4f %s = %.4f %s\n", amount, valuteFrom, result, valuteTo)


	// fmt.Println(string(body))

	// router := gin.Default()
	// router.GET("/albums", getAlbums)
	// router.Run("localhost:8080")
	// fmt.Println("server is running on port 8080weweq")
}
