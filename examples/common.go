package main

type Order struct {
	OrderID    int
	CustomerID string
	Status     string
}

const (
	streamName = "test_stream"
	subject = "test_subject"
)
