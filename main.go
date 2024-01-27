package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"context"
	"time"

	"github.com/gocolly/colly"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"

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

		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(searchTerm)) {
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

		if item.Title != "" && strings.Contains(strings.ToLower(item.Title), strings.ToLower(searchTerm)) {
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

	// Vai gerar uma context diferente para cada website a ser acessado pelo chromedp
	saoVicenteContext, cancel := chromedp.NewContext(context.Background())
	// pavanContext, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Scrape the DIV content from web js rendered websites
	var saoVicenteNodes []*cdp.Node
	// var pavanNodes []*cdp.Node

	// Vai carregar o código renderizado via javascript do website
	chromedp.Run(saoVicenteContext,
		chromedp.Navigate("https://www.sitemercado.com.br/supermercadossaovicente/santa-barbara-d-oeste-loja-sao-vicente-centro-sbo-centro-graca-martins"),
		chromedp.Sleep(10*time.Second),
		chromedp.Nodes(".list-product-item", &saoVicenteNodes, chromedp.ByQueryAll),
	)

	for _, node := range saoVicenteNodes {

		var photo, name, price string

		// Vai realizar um laço de repetição para cada produto da loja
		chromedp.Run(saoVicenteContext,
			chromedp.AttributeValue("img", "src", &photo, nil, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.AttributeValue(".list-product-link", "aria-label", &name, nil, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(".preco-oferta", &price, chromedp.ByQuery, chromedp.FromNode(node)),
		)

		// Geração do item e adição do mesmo no objeto de produtos
		product := item{
			Unidade: "São Vicente",
			Photo: photo,
			Title: name,
			Price: price,
		}

		items = append(items, product)
	}

	// Vai carregar o código renderizado via javascript do website
	// chromedp.Run(pavanContext,
	// 	chromedp.Navigate("https://www.sitemercado.com.br/supermercadospavan/santa-barbara-d-oeste-loja-jardim-perola-jardim-esmeralda-av-do-comercio"),
	// 	chromedp.Nodes(".list-product-item", &pavanNodes, chromedp.ByQueryAll),
	// )

	// for _, node := range pavanNodes {

	// 	var photo, name, price string

	// 	// Vai realizar um laço de repetição para cada produto da loja
	// 	chromedp.Run(pavanContext,
	// 		chromedp.AttributeValue("img", "src", &photo, nil, chromedp.ByQuery, chromedp.FromNode(node)),
	// 		chromedp.AttributeValue(".list-product-link", "aria-label", &name, nil, chromedp.ByQuery, chromedp.FromNode(node)),
	// 		chromedp.Text(".preco-oferta", &price, chromedp.ByQuery, chromedp.FromNode(node)),
	// 	)

	// 	// Geração do item e adição do mesmo no objeto de produtos
	// 	product := item{
	// 		Unidade: "Pavan",
	// 		Photo: photo,
	// 		Title: name,
	// 		Price: price,
	// 	}

	// 	items = append(items, product)
	// }

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

	// Configurações dos middlewares de conexão da API
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
