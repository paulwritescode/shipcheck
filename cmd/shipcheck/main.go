// Command shipcheck is the ShipCheck backend entrypoint.
//
// The handler logic lives in internal/api; this binary wires it up. When
// deployed, it is built as a static binary named "bootstrap" and runs on the
// AWS Lambda provided.al2023 custom runtime behind an API Gateway HTTP API,
// backed by DynamoDB. Locally it runs as a plain HTTP server backed by an
// in-memory repository, so no AWS connection is required to work on ShipCheck.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/paulwritescode/shipcheck/internal/api"
	"github.com/paulwritescode/shipcheck/internal/repo"
	"github.com/paulwritescode/shipcheck/internal/seed"
)

func main() {
	ctx := context.Background()
	repository := newRepository(ctx)

	// Register the seeded "Team Inbox 2.0" demo launch so a visitor can explore
	// it with no account. Seeding is idempotent (fixed id), so it is safe to run
	// on every start, including against DynamoDB.
	api.SetSeedBuilder(seed.TeamInbox)
	if _, err := repository.Create(ctx, seed.TeamInbox()); err != nil {
		log.Printf("warning: could not seed demo launch: %v", err)
	}

	handler := api.NewRouter(repository)

	// Lambda detection: the custom runtime sets this environment variable.
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		api.ServeLambda(handler)
		return
	}

	addr := os.Getenv("SHIPCHECK_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("ShipCheck dev server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// newRepository selects the persistence backend. If SHIPCHECK_TABLE is set
// (deployment), it uses DynamoDB with the ambient AWS config (the Lambda
// execution role). Otherwise it uses the in-memory repository (local/dev).
func newRepository(ctx context.Context) repo.LaunchRepository {
	table := os.Getenv("SHIPCHECK_TABLE")
	if table == "" {
		return repo.NewInMemoryLaunchRepository()
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("loading AWS config: %v", err)
	}
	shareIndex := os.Getenv("SHIPCHECK_SHARE_INDEX")
	if shareIndex == "" {
		shareIndex = "share-index"
	}
	return repo.NewDynamoDBLaunchRepository(dynamodb.NewFromConfig(cfg), table, shareIndex)
}
