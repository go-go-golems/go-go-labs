package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type StartCommand struct{}
type UploadPDFCommand struct{}
type StartTextractCommand struct {
	PDFLocation string
}

type PDFUploadedEvent struct {
	PDFLocation string
}
type TextractStartedEvent struct {
	PDFLocation string
}
type OCRFinishedEvent struct {
	CSVLocation string
}

func main() {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewStdLogger(false, false))
	publisher := pubSub
	subscriber := pubSub

	ctx := context.Background()

	// Goroutine simulating the entire workflow based on commands and events
	go func() {
		uploadPDFSubscription, _ := subscriber.Subscribe(ctx, "upload_pdf_command")
		startTextractSubscription, _ := subscriber.Subscribe(ctx, "start_textract_command")

		for {
			select {
			case <-uploadPDFSubscription:
				time.Sleep(2 * time.Second) // Simulate PDF upload delay
				pdfLocation := "mocked_pdf_location"
				_ = publisher.Publish("pdf_uploaded_event", message.NewMessage(watermill.NewUUID(), []byte(pdfLocation)))

			case msg := <-startTextractSubscription:
				time.Sleep(3 * time.Second) // Simulate Textract processing delay
				pdfLocation := string(msg.Payload)
				_ = pdfLocation
				csvLocation := "mocked_csv_location"
				_ = publisher.Publish("ocr_finished_event", message.NewMessage(watermill.NewUUID(), []byte(csvLocation)))
			}
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Goroutine displaying progress based on events
	go func() {
		pdfUploadedSubscription, _ := subscriber.Subscribe(ctx, "pdf_uploaded_event")
		textractStartedSubscription, _ := subscriber.Subscribe(ctx, "textract_started_event")
		ocrFinishedSubscription, _ := subscriber.Subscribe(ctx, "ocr_finished_event")

		for {
			select {
			case <-pdfUploadedSubscription:
				fmt.Println("PDF uploaded!")
			case <-textractStartedSubscription:
				fmt.Println("Textract processing started!")
			case <-ocrFinishedSubscription:
				fmt.Println("OCR finished! CSV available at mocked_csv_location")
				wg.Done()
				return
			}
		}
	}()

	// Goroutine simulating the handler that starts the Textract job upon PDF upload
	go func() {
		pdfUploadedSubscription, _ := subscriber.Subscribe(ctx, "pdf_uploaded_event")

		for {
			msg := <-pdfUploadedSubscription
			pdfLocation := string(msg.Payload)
			fmt.Println("Received PDF uploaded event. Starting Textract job...")
			_ = publisher.Publish("start_textract_command", message.NewMessage(watermill.NewUUID(), []byte(pdfLocation)))
		}
	}()

	time.Sleep(1 * time.Second)
	fmt.Println("Starting workflow...")
	// Start the workflow
	_ = publisher.Publish("upload_pdf_command", message.NewMessage(watermill.NewUUID(), nil))

	wg.Wait()
}
