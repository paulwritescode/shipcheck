package repo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// DynamoDBLaunchRepository persists launches in a single Amazon DynamoDB table.
//
// Table design (declared in the CDK stack, not here):
//   - Partition key: "pk" (the launch id), one item per launch.
//   - GSI "share-index": partition key "shareToken", used only by
//     GetByShareToken. The shareToken attribute is present only when a launch
//     has an enabled share link, so disabled/absent links are not indexed.
//
// The launch aggregate (items, risks, evidence, brief, share link) is stored
// as a marshaled attribute map under "data". The assessment is recomputed on
// read and on every mutation rather than persisted, keeping DynamoDB the
// system of record for inputs and the pure engine the sole authority for the
// derived status.
type DynamoDBLaunchRepository struct {
	client   DynamoAPI
	table    string
	shareIdx string
}

// DynamoAPI is the subset of the DynamoDB client this repository uses. Narrowing
// to an interface keeps the repository unit-testable with a fake and avoids a
// hard dependency on a live client.
type DynamoAPI interface {
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	DeleteItem(ctx context.Context, in *dynamodb.DeleteItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, in *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// NewDynamoDBLaunchRepository builds a repository over the given client/table.
func NewDynamoDBLaunchRepository(client DynamoAPI, table, shareIndex string) *DynamoDBLaunchRepository {
	return &DynamoDBLaunchRepository{client: client, table: table, shareIdx: shareIndex}
}

var _ LaunchRepository = (*DynamoDBLaunchRepository)(nil)

// item is the stored shape: the launch id key, the optional share-token GSI
// key, and the marshaled launch data.
type item struct {
	PK         string            `dynamodbav:"pk"`
	ShareToken string            `dynamodbav:"shareToken,omitempty"`
	Data       domain.LaunchData `dynamodbav:"data"`
}

func (r *DynamoDBLaunchRepository) put(ctx context.Context, data domain.LaunchData) error {
	it := item{PK: data.ID, Data: cloneLaunchData(data)}
	if data.Share != nil && data.Share.Enabled {
		it.ShareToken = data.Share.Token
	}
	av, err := attributevalue.MarshalMap(it)
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      av,
	})
	return err
}

func (r *DynamoDBLaunchRepository) load(ctx context.Context, id string) (domain.LaunchData, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key:       map[string]ddbtypes.AttributeValue{"pk": &ddbtypes.AttributeValueMemberS{Value: id}},
	})
	if err != nil {
		return domain.LaunchData{}, err
	}
	if out.Item == nil {
		return domain.LaunchData{}, ErrNotFound
	}
	var it item
	if err := attributevalue.UnmarshalMap(out.Item, &it); err != nil {
		return domain.LaunchData{}, err
	}
	return it.Data, nil
}

func (r *DynamoDBLaunchRepository) Create(ctx context.Context, input NewLaunch) (domain.Launch, error) {
	id := input.ID
	if id == "" {
		id = newID()
	}
	data := domain.LaunchData{
		ID: id, Name: input.Name, Description: input.Description, TargetDate: input.TargetDate,
		Owner: input.Owner, ProductArea: input.ProductArea, RepositoryURL: input.RepositoryURL,
		DocumentationURL: input.DocumentationURL, Brief: input.Brief,
		ChecklistItems: input.ChecklistItems, Risks: input.Risks,
	}
	if err := r.put(ctx, data); err != nil {
		return domain.Launch{}, err
	}
	return assess(data), nil
}

func (r *DynamoDBLaunchRepository) Get(ctx context.Context, id string) (domain.Launch, error) {
	data, err := r.load(ctx, id)
	if err != nil {
		return domain.Launch{}, err
	}
	return assess(data), nil
}

func (r *DynamoDBLaunchRepository) GetByShareToken(ctx context.Context, token string) (domain.Launch, error) {
	if token == "" {
		return domain.Launch{}, ErrNotFound
	}
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String(r.shareIdx),
		KeyConditionExpression: aws.String("shareToken = :t"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":t": &ddbtypes.AttributeValueMemberS{Value: token},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return domain.Launch{}, err
	}
	if len(out.Items) == 0 {
		return domain.Launch{}, ErrNotFound
	}
	var it item
	if err := attributevalue.UnmarshalMap(out.Items[0], &it); err != nil {
		return domain.Launch{}, err
	}
	// Defensively confirm the link is still enabled and matches.
	if it.Data.Share == nil || !it.Data.Share.Enabled || it.Data.Share.Token != token {
		return domain.Launch{}, ErrNotFound
	}
	return assess(it.Data), nil
}

func (r *DynamoDBLaunchRepository) Update(ctx context.Context, data domain.LaunchData) (domain.Launch, error) {
	if _, err := r.load(ctx, data.ID); err != nil {
		return domain.Launch{}, err
	}
	if err := r.put(ctx, data); err != nil {
		return domain.Launch{}, err
	}
	return assess(data), nil
}

func (r *DynamoDBLaunchRepository) mutate(ctx context.Context, launchID string, fn func(*domain.LaunchData)) (domain.Launch, error) {
	data, err := r.load(ctx, launchID)
	if err != nil {
		return domain.Launch{}, err
	}
	working := cloneLaunchData(data)
	fn(&working)
	if err := r.put(ctx, working); err != nil {
		return domain.Launch{}, err
	}
	return assess(working), nil
}

func (r *DynamoDBLaunchRepository) AddChecklistItem(ctx context.Context, launchID string, i domain.ChecklistItem) (domain.Launch, error) {
	return r.mutate(ctx, launchID, func(d *domain.LaunchData) { d.ChecklistItems = append(d.ChecklistItems, i) })
}

func (r *DynamoDBLaunchRepository) UpdateChecklistItem(ctx context.Context, launchID string, i domain.ChecklistItem) (domain.Launch, error) {
	return r.mutate(ctx, launchID, func(d *domain.LaunchData) {
		for idx := range d.ChecklistItems {
			if d.ChecklistItems[idx].ID == i.ID {
				d.ChecklistItems[idx] = i
				return
			}
		}
	})
}

func (r *DynamoDBLaunchRepository) DeleteChecklistItem(ctx context.Context, launchID, itemID string) (domain.Launch, error) {
	return r.mutate(ctx, launchID, func(d *domain.LaunchData) {
		out := d.ChecklistItems[:0]
		for _, it := range d.ChecklistItems {
			if it.ID != itemID {
				out = append(out, it)
			}
		}
		d.ChecklistItems = out
	})
}

func (r *DynamoDBLaunchRepository) AddRisk(ctx context.Context, launchID string, rk domain.Risk) (domain.Launch, error) {
	return r.mutate(ctx, launchID, func(d *domain.LaunchData) { d.Risks = append(d.Risks, rk) })
}

func (r *DynamoDBLaunchRepository) UpdateRisk(ctx context.Context, launchID string, rk domain.Risk) (domain.Launch, error) {
	return r.mutate(ctx, launchID, func(d *domain.LaunchData) {
		for idx := range d.Risks {
			if d.Risks[idx].ID == rk.ID {
				d.Risks[idx] = rk
				return
			}
		}
	})
}

func (r *DynamoDBLaunchRepository) DeleteRisk(ctx context.Context, launchID, riskID string) (domain.Launch, error) {
	return r.mutate(ctx, launchID, func(d *domain.LaunchData) {
		out := d.Risks[:0]
		for _, rk := range d.Risks {
			if rk.ID != riskID {
				out = append(out, rk)
			}
		}
		d.Risks = out
	})
}

func (r *DynamoDBLaunchRepository) SetShareLink(ctx context.Context, launchID string, link *domain.ShareLink) (domain.Launch, error) {
	return r.mutate(ctx, launchID, func(d *domain.LaunchData) { d.Share = link })
}

func (r *DynamoDBLaunchRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.table),
		Key:       map[string]ddbtypes.AttributeValue{"pk": &ddbtypes.AttributeValueMemberS{Value: id}},
	})
	return err
}
