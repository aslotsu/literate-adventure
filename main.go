package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aslotsu/notification-worker/config"
	"github.com/aslotsu/notification-worker/handlers"
	"github.com/nats-io/nats.go"
)

func main() {
	log.Println("🚀 Exobook Notification Worker Starting...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log.Printf("✅ Configuration loaded (env=%s, region=%s)", cfg.Environment, cfg.AWSRegion)

	// Connect to NATS
	log.Printf("📡 Connecting to NATS at %s...", cfg.NatsURL)

	var nc *nats.Conn
	if cfg.NatsCredsFile != "" {
		// Connect with credentials file
		nc, err = nats.Connect(
			cfg.NatsURL,
			nats.UserCredentials(cfg.NatsCredsFile),
			nats.Name("notification-worker"),
			nats.ReconnectWait(nats.DefaultReconnectWait),
			nats.MaxReconnects(-1), // Unlimited reconnects
		)
	} else {
		// Connect without credentials (for local dev)
		nc, err = nats.Connect(
			cfg.NatsURL,
			nats.Name("notification-worker"),
		)
	}

	if err != nil {
		log.Fatalf("❌ Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	log.Println("✅ Connected to NATS")

	// Setup connection status callbacks
	nc.SetDisconnectErrHandler(func(nc *nats.Conn, err error) {
		if err != nil {
			log.Printf("⚠️  Disconnected from NATS: %v", err)
		}
	})

	nc.SetReconnectHandler(func(nc *nats.Conn) {
		log.Printf("🔄 Reconnected to NATS at %s", nc.ConnectedUrl())
	})

	nc.SetClosedHandler(func(nc *nats.Conn) {
		log.Println("🔌 NATS connection closed")
	})

	// Initialize notification service
	log.Printf("💾 Initializing DynamoDB notification service (table=%s)...", cfg.NotifTableName)

	notifService, err := handlers.NewNotificationService(
		cfg.AWSRegion,
		cfg.NotifTableName,
		cfg.PusherAppID,
		cfg.PusherKey,
		cfg.PusherSecret,
		cfg.PusherCluster,
	)
	if err != nil {
		log.Fatalf("❌ Failed to initialize notification service: %v", err)
	}

	log.Println("✅ Notification service initialized")

	// Create and start worker
	worker := handlers.NewNotificationWorker(nc, notifService)

	if err := worker.Start(); err != nil {
		log.Fatalf("❌ Failed to start worker: %v", err)
	}

	log.Println("🎉 Notification worker is running!")
	log.Println("📬 Listening for events on: notifications.>")
	log.Println("   - notifications.post.like")
	log.Println("   - notifications.post.unlike")
	log.Println("   - notifications.comment.like")
	log.Println("   - notifications.reply.post")
	log.Println("   - notifications.reply.comment")
	log.Println()
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Block until signal received
	<-sigCh

	log.Println()
	log.Println("🛑 Shutdown signal received, cleaning up...")

	// Graceful shutdown
	if err := worker.Stop(); err != nil {
		log.Printf("⚠️  Error stopping worker: %v", err)
	}

	log.Println("👋 Notification worker stopped")
}
