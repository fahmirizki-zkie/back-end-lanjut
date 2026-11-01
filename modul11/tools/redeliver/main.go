package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"modul11/app/model"
	"modul11/config"
	"modul11/database"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID outbox event")
	copies := flag.Int("copies", 5, "Jumlah salinan task yang di-enqueue")
	flag.Parse()

	if *outboxID <= 0 {
		log.Fatal("flag -outbox wajib diisi angka positif")
	}

	config.LoadEnv()

	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}
	defer pool.Close()

	var event model.OutboxEvent
	err = pool.QueryRow(ctx, `
		SELECT id, event_type, payload
		FROM outbox
		WHERE id = $1`, *outboxID).Scan(&event.ID, &event.EventType, &event.Payload)
	if err != nil {
		log.Fatalf("gagal mengambil outbox ID %d: %v", *outboxID, err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	job := model.NotificationJob{
		OutboxID:  event.ID,
		EventType: event.EventType,
		Payload:   event.Payload,
	}

	payloadBytes, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("gagal marshal job: %v", err)
	}

	for i := 1; i <= *copies; i++ {
		task := asynq.NewTask(
			model.TaskSendNotification,
			payloadBytes,
			asynq.Queue(model.QueueNotifications),
			asynq.TaskID(fmt.Sprintf("redeliver-%d-%d", event.ID, i)),
		)

		_, err := client.EnqueueContext(ctx, task)
		if err != nil {
			log.Fatalf("gagal enqueue copy %d: %v", i, err)
		}
	}

	fmt.Printf("Berhasil melakukan enqueue %d salinan task untuk outbox ID %d ke Redis\n", *copies, event.ID)
}
