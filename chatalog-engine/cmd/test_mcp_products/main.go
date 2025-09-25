package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/defryfazz/fazztalog/internal/product"
)

func main() {
	ctx := context.Background()

	svc, err := product.NewMCPService(ctx)
	if err != nil {
		log.Fatal("init MCP:", err)
	}
	defer svc.Close()

	phone := os.Getenv("TEST_PHONE")
	if phone == "" {
		phone = "085224416325"
	} // contoh default

	out, err := svc.ListByPhone(ctx, phone, 20, 0)
	if err != nil {
		log.Fatal("list products:", err)
	}

	fmt.Printf("Total: %d\n", out.Total)
	for i, p := range out.Items {
		fmt.Printf("%2d. %s — %.2f %s\n", i+1, p.Name, p.Price, p.Currency)
	}
	if out.NextOffset != nil {
		fmt.Println("Next offset:", *out.NextOffset)
	}
}
