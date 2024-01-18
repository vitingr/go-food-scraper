package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"

	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type item struct {
	Photo       string `json:"photo"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       string `json:"price"`
	IsOffer     bool   `json:"isOffer"`
	OfferPrice  string `json:"offerPrice"`
}

// func NewServer(c *mongo.Client) *Server {
// 	return &Server {
// 		client: c
// 	}
// }

func handleGetSupermarketsData(w http.ResponseWriter, r *http.Request) {
	data := getSupermarketData()

	// Response do Servidor
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)

}

func getSupermarketData() []item {
	call := colly.NewCollector(
		colly.AllowedDomains("https://www.higa.com.br/", "www.higa.com.br", "https://atacadao.com.br/catalogo", "www.atacadao.com.br/catalogo", "https://www.superpaguemenos.com.br/", "www.superpaguemenos.com.br"),
	)

	var items []item

	call.OnHTML("div.item-product", func(h *colly.HTMLElement) {

		// h.ForEach("div.card-body", func(_ int, e *colly.HTMLElement) {
			item := item{
				Photo: h.ChildAttr("img", "data-src"),
				Title: h.ChildText("h2.title"),
				Price: h.ChildText("p.sale-price"),
			}

		// 	item.Title = strings.ReplaceAll(item.Title, "\n", "")
		// 	item.Price = strings.ReplaceAll(item.Price, "\t", "")

			items = append(items, item)
		// })
	})

	call.OnRequest(func(r *colly.Request) {
		fmt.Println(r.URL.String())
	})

	err := call.Visit("https://www.superpaguemenos.com.br/")
	if err != nil {
		log.Fatal(err)
	}

	higaContent, err := json.Marshal(items)

	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile("higa.json", higaContent, 0644)

	return items
}

func main() {

	router := mux.NewRouter()
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"*"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)
	router.Use(corsMiddleware)

	// Get Supermarket Products Data
	router.HandleFunc("/data", handleGetSupermarketsData).Methods("GET")

	http.ListenAndServe(":3030", router)
}
