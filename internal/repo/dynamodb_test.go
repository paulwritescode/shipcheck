package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// fakeDynamo is an in-memory stand-in for the DynamoDB client that stores the
// real marshaled attribute maps. It exercises the repository's actual
// marshal/unmarshal and share-token query logic without AWS or DynamoDB Local,
// so these tests run anywhere. A full integration test against DynamoDB Local
// can be added when running with AWS, but is not required to verify the mapping.
type fakeDynamo struct {
	items map[string]map[string]ddbtypes.AttributeValue // pk -> item
}

func newFakeDynamo() *fakeDynamo {
	return &fakeDynamo{items: map[string]map[string]ddbtypes.AttributeValue{}}
}

func (f *fakeDynamo) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	pk := in.Item["pk"].(*ddbtypes.AttributeValueMemberS).Value
	f.items[pk] = in.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (f *fakeDynamo) GetItem(_ context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	pk := in.Key["pk"].(*ddbtypes.AttributeValueMemberS).Value
	item, ok := f.items[pk]
	if !ok {
		return &dynamodb.GetItemOutput{}, nil // nil Item => not found
	}
	return &dynamodb.GetItemOutput{Item: item}, nil
}

func (f *fakeDynamo) DeleteItem(_ context.Context, in *dynamodb.DeleteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	pk := in.Key["pk"].(*ddbtypes.AttributeValueMemberS).Value
	delete(f.items, pk)
	return &dynamodb.DeleteItemOutput{}, nil
}

func (f *fakeDynamo) Query(_ context.Context, in *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	// Emulate the share-index query: match items whose shareToken equals :t.
	want := in.ExpressionAttributeValues[":t"].(*ddbtypes.AttributeValueMemberS).Value
	var out []map[string]ddbtypes.AttributeValue
	for _, item := range f.items {
		if tok, ok := item["shareToken"].(*ddbtypes.AttributeValueMemberS); ok && tok.Value == want {
			out = append(out, item)
		}
	}
	return &dynamodb.QueryOutput{Items: out}, nil
}

func newDDBRepo() (*DynamoDBLaunchRepository, context.Context) {
	return NewDynamoDBLaunchRepository(newFakeDynamo(), "shipcheck", "share-index"), context.Background()
}

func TestDynamoRoundTrip(t *testing.T) {
	r, ctx := newDDBRepo()
	created, err := r.Create(ctx, NewLaunch{
		ID: "l1", Name: "Team Inbox", TargetDate: "2026-10-16", Owner: "Alice",
		ChecklistItems: []domain.ChecklistItem{{
			ID: "i1", Title: "migrate", Category: domain.CategoryEngineering,
			Priority: domain.PriorityHigh, CompletionState: domain.CompletionComplete,
			IsCritical: true, Evidence: []domain.Evidence{{ID: "e1", URL: "https://example.com/pr/1"}},
		}},
		Risks: []domain.Risk{{ID: "r1", Title: "scale", Severity: domain.SeverityHigh, Status: domain.RiskResolved}},
	})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := r.Get(ctx, "l1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != created.Name || len(loaded.ChecklistItems) != 1 || len(loaded.Risks) != 1 {
		t.Fatalf("round-trip lost data: %+v", loaded.LaunchData)
	}
	if loaded.ChecklistItems[0].Evidence[0].URL != "https://example.com/pr/1" {
		t.Fatalf("evidence url lost: %+v", loaded.ChecklistItems[0].Evidence)
	}
	// Assessment is recomputed on read (resolved high-severity risk is not blocking;
	// the critical item is complete with evidence) -> Ready.
	if loaded.Assessment.Status != domain.StatusReady {
		t.Fatalf("want Ready, got %s", loaded.Assessment.Status)
	}
}

func TestDynamoGetNotFound(t *testing.T) {
	r, ctx := newDDBRepo()
	if _, err := r.Get(ctx, "nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestDynamoShareTokenQuery(t *testing.T) {
	r, ctx := newDDBRepo()
	if _, err := r.Create(ctx, NewLaunch{ID: "l1", Name: "n", TargetDate: "2026-10-16", Owner: "o"}); err != nil {
		t.Fatal(err)
	}
	// Enable a share link; it should become queryable via the index.
	if _, err := r.SetShareLink(ctx, "l1", &domain.ShareLink{Token: "abc123", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	got, err := r.GetByShareToken(ctx, "abc123")
	if err != nil {
		t.Fatalf("token should resolve: %v", err)
	}
	if got.ID != "l1" {
		t.Fatalf("wrong launch %s", got.ID)
	}
	// Disable -> shareToken attribute dropped -> no longer indexed/resolvable.
	if _, err := r.SetShareLink(ctx, "l1", &domain.ShareLink{Token: "abc123", Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetByShareToken(ctx, "abc123"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("disabled token should not resolve, got %v", err)
	}
}
