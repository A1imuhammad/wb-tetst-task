package models

import (
	"time"
)


type KafkaConfig struct {
    KAFKA_BROKER string `yaml:"KKAFKA_BROKER"`
    KAFKA_TOPIC   string `yaml:"KAFKA_TOPIC"`
    KAFKA_GROUP   string `yaml:"KAFKA_GROUP"`
}

type PostgresConfig struct {
	POSTGRES_USER string `yaml:"POSTGRES_USER"`
	POSTGRES_PASSWORD string `yaml:"POSTGRES_PASSWORD"`
	POSTGRES_PORT uint16 `yaml:"POSTGRES_PORT"`
	POSTGRES_DB string `yaml:"POSTGRES_DB"`
	POSTGRES_HOST string `yaml:"POSTGRES_HOST"`

}

type HttpServerConfig struct {
	Address     string        `yaml:"ADDRESS"`
	Timeout     time.Duration `yaml:"TIMEOUT"`
}


type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`

	Delivery Delivery   `json:"delivery"`
	Payment  Payment    `json:"payment"`
	Items    []Item     `json:"items"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int64  `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int64  `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}
