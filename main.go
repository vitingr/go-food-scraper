package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"

	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type item struct {
	Unidade     string `json:"unidade"`
	Photo       string `json:"photo"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       string `json:"price"`
	IsOffer     bool   `json:"isOffer"`
	OfferPrice  string `json:"offerPrice"`
}

func handleGetSupermarketsData(w http.ResponseWriter, r *http.Request) {
	data := getSupermarketData()

	// Response do Servidor
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)
}

func handleSearchSupermarketData(w http.ResponseWriter, r *http.Request) {
	param := mux.Vars(r)
	searchTerm, ok := param["searchTerm"]

	if !ok {
		http.Error(w, "SearchTerm parameter is missing", http.StatusBadRequest)
		return
	}

	data := getSearchData(searchTerm)

	// Response do Servidor
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)
}

func getSearchData(searchTerm string) []item {
	var items []item

	callPagueMenos := colly.NewCollector(
		colly.AllowedDomains("https://atacadao.com.br/catalogo", "www.atacadao.com.br/catalogo", "https://www.superpaguemenos.com.br/", "www.superpaguemenos.com.br"),
	)

	callHiga := colly.NewCollector(
		colly.AllowedDomains("https://www.higa.com.br/", "www.higa.com.br"),
	)

	callPagueMenos.OnHTML("div.item-product", func(h *colly.HTMLElement) {

		item := item{
			Unidade:    "Pague menos",
			Photo:      h.ChildAttr("img", "data-src"),
			Title:      h.ChildText("h2.title"),
			Price:      h.ChildText("p.unit-price"),
			OfferPrice: h.ChildText("p.sale-price"),
		}

		if strings.Contains(item.Title, searchTerm) {
			items = append(items, item)
		}
	})

	callHiga.OnHTML("div.swiper-slide", func(h *colly.HTMLElement) {
		item := item{
			Unidade: "Higa",
			Photo:   h.ChildAttr("img.produto-img", "src"),
			Title:   h.ChildText("h3.text-muted"),
			Price:   h.ChildText("span.fw-bolder"),
		}

		if item.Title != "" && strings.Contains(item.Title, searchTerm) {
			items = append(items, item)
		}
	})

	callPagueMenos.OnRequest(func(r *colly.Request) {
		fmt.Println(r.URL.String())
	})

	callHiga.OnRequest(func(r *colly.Request) {
		fmt.Println(r.URL.String())
	})

	err := callPagueMenos.Visit("https://www.superpaguemenos.com.br/")
	if err != nil {
		log.Fatal(err)
	}

	err = callHiga.Visit("https://www.higa.com.br/")
	if err != nil {
		log.Fatal(err)
	}

	productsContent, err := json.Marshal(items)

	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile("data.json", productsContent, 0644)

	return items
}

func getSupermarketData() []item {

	var items []item

	callPagueMenos := colly.NewCollector(
		colly.AllowedDomains("https://atacadao.com.br/catalogo", "www.atacadao.com.br/catalogo", "https://www.superpaguemenos.com.br/", "www.superpaguemenos.com.br"),
	)

	callHiga := colly.NewCollector(
		colly.AllowedDomains("https://www.higa.com.br/", "www.higa.com.br"),
	)

	callPagueMenos.OnHTML("div.item-product", func(h *colly.HTMLElement) {

		item := item{
			Unidade:    "Pague menos",
			Photo:      h.ChildAttr("img", "data-src"),
			Title:      h.ChildText("h2.title"),
			Price:      h.ChildText("p.unit-price"),
			OfferPrice: h.ChildText("p.sale-price"),
		}

		items = append(items, item)
	})

	callHiga.OnHTML("div.swiper-slide", func(h *colly.HTMLElement) {
		item := item{
			Unidade: "Higa",
			Photo:   h.ChildAttr("img.produto-img", "src"),
			Title:   h.ChildText("h3.text-muted"),
			Price:   h.ChildText("span.fw-bolder"),
		}

		if item.Title != "" {
			items = append(items, item)
		}
	})

	callPagueMenos.OnRequest(func(r *colly.Request) {
		fmt.Println(r.URL.String())
	})

	callHiga.OnRequest(func(r *colly.Request) {
		fmt.Println(r.URL.String())
	})

	err := callPagueMenos.Visit("https://www.superpaguemenos.com.br/")
	if err != nil {
		log.Fatal(err)
	}

	err = callHiga.Visit("https://www.higa.com.br/")
	if err != nil {
		log.Fatal(err)
	}

	productsContent, err := json.Marshal(items)

	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile("data.json", productsContent, 0644)

	return items
}

func main() {

	// Configuração de Rotas
	router := mux.NewRouter()

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"*"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	router.Use(corsMiddleware)

	// Get Supermarket Products Data
	router.HandleFunc("/data", handleGetSupermarketsData).Methods("GET")

	// Get SearchTerm Products Data
	router.HandleFunc("/data/{searchTerm}", handleSearchSupermarketData).Methods("GET")

	http.ListenAndServe(":3030", router)
}
