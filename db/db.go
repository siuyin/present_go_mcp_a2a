// Package db provide a demo inventory database.
package db

import (
	"fmt"
	"strings"
)

type dbEntry struct {
	ID          string `json:"id"`
	ProductName string `json:"product_name"`
	PriceUSD    string `json:"price_usd"`
	QtyInStock  int    `json:"quantity_in_stock"`
}

func (d dbEntry) String() string {
	return fmt.Sprintf("ID=%s, product name=%s, price in USD=%s, quantity in stock=%d", d.ID, d.ProductName, d.PriceUSD, d.QtyInStock)
}

var inventory map[string]dbEntry

func init() {
	inventory = make(map[string]dbEntry)
	loadData()
}

func loadData() {
	dat := []dbEntry{
		dbEntry{"1", "iPhone 14", "899", 0},
		dbEntry{"2", "iPhone 15", "1,099", 56},
		dbEntry{"3", "simpleX", "199", 32},
	}
	for _, d := range dat {
		inventory[strings.ToLower(d.ProductName)] = d
	}
}

func Get(productName string) string {
	item, found := inventory[strings.ToLower(productName)]
	if !found {
		return fmt.Sprintf("Sorry we do not have any %s", productName)
	}
	return item.String()
}

func keysOfMap[T any](m map[string]T) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func List() []string {
	return keysOfMap(inventory)
}
