package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

const (
	ORDERS    = "orders"
	CUSTOMERS = "customers"
	PRODUCTS  = "products"
)

type DashboardData struct {
	Orders    int `json:"orders"`
	Customers int `json:"customers"`
	Products  int `json:"products"`
}

var ChannelMap map[*websocket.Conn]chan int

func (dashboard *DashboardData) FetchDashboardHelper() []byte {
	data, err := json.Marshal(dashboard)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	return data
}

func (dashboard *DashboardData) AddDashboardData(s string) {
	switch s {
	case ORDERS:
		dashboard.Orders++
	case CUSTOMERS:
		dashboard.Customers++
	case PRODUCTS:
		dashboard.Products++
	}
	// update API
	for _, element := range ChannelMap {
		element <- 1
	}
	return
}

func (dashboard *DashboardData) RemoveDashboardData(s string) error {
	switch s {
	case ORDERS:
		if dashboard.Orders == 0 {
			return fmt.Errorf("no order remains to remove")
		}
		dashboard.Orders--
	case CUSTOMERS:
		if dashboard.Customers == 0 {
			return fmt.Errorf("no customer remains to remove")
		}
		dashboard.Customers--
	case PRODUCTS:
		if dashboard.Products == 0 {
			return fmt.Errorf("no products remains to remove")
		}
		dashboard.Products--
	}
	// update API
	for _, element := range ChannelMap {
		element <- 1
	}
	return nil
}

var dashboard DashboardData
var upgrages = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	// ChannelMap = make(map[*websocket.Conn]chan int)
	fmt.Println("Starting server ...")
	router := chi.NewRouter()

	router.Route("/", func(ws chi.Router) {
		ws.Get("/", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("<h4>Welcome to Dashboard app</h4>"))
			if err != nil {
				return
			}
			log.Printf("welcome to dashboard app")
		})
		ws.Get("/dashboard", DashboardHandler)
		ws.Post("/sign-up", AddCustomer)
		ws.Post("/order", AddOrder)
		ws.Post("/product", AddProduct)
		ws.Delete("/order", RemoveOrder)
		ws.Delete("/product", RemoveProduct)
		ws.Delete("/sign-off", RemoveCustomer)
	})
	log.Fatal(http.ListenAndServe(":8082", router))

}
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	update := make(chan int, 1)
	upgrages.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrages.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	var data []byte
	data = dashboard.FetchDashboardHelper()
	conn.WriteMessage(1, data)

	ChannelMap[conn] = update
	func(conn *websocket.Conn, update chan int) {
		for {
			<-update
			data = dashboard.FetchDashboardHelper()
			conn.WriteMessage(1, data)
		}
	}(conn, ChannelMap[conn])
	delete(ChannelMap, conn)
	return
}

func AddOrder(w http.ResponseWriter, r *http.Request) {
	dashboard.AddDashboardData(ORDERS)
	_, err := w.Write([]byte("Order Added"))
	if err != nil {
		return
	}
	return
}

func AddCustomer(w http.ResponseWriter, r *http.Request) {
	dashboard.AddDashboardData(CUSTOMERS)
	_, err := w.Write([]byte("Customer Added"))
	if err != nil {
		return
	}
	return
}

func AddProduct(w http.ResponseWriter, r *http.Request) {
	dashboard.AddDashboardData(PRODUCTS)
	_, err := w.Write([]byte("Product Added"))
	if err != nil {
		return
	}
	return
}

func RemoveOrder(w http.ResponseWriter, r *http.Request) {
	if dashboard.Orders == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := dashboard.RemoveDashboardData(ORDERS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = w.Write([]byte("Order Removed"))
	if err != nil {
		return
	}
	return
}

func RemoveCustomer(w http.ResponseWriter, r *http.Request) {
	if dashboard.Customers == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := dashboard.RemoveDashboardData(CUSTOMERS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = w.Write([]byte("Customer Removed"))
	if err != nil {
		return
	}
	return
}

func RemoveProduct(w http.ResponseWriter, r *http.Request) {
	if dashboard.Products == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := dashboard.RemoveDashboardData(PRODUCTS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = w.Write([]byte("Product Removed"))
	if err != nil {
		return
	}
	return
}
