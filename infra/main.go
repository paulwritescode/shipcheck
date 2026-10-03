// Command infra is the AWS CDK application for ShipCheck, authored in Go so the
// whole project stays in one language.
//
// It provisions an always-free-tier serverless stack:
//   - a Go Lambda (provided.al2023, arm64) running the ShipCheck backend,
//   - an API Gateway HTTP API routing /api/* to the Lambda,
//   - a DynamoDB table (pk partition key) with a "share-index" GSI on
//     shareToken for public-report resolution,
//   - an S3 bucket + CloudFront distribution serving the React SPA over HTTPS,
//     with the SPA calling /api/* through CloudFront to the API.
//
// The deployed public demo defaults to the free deterministic advisor, so no
// model key is required. If an opt-in provider is enabled, its key is supplied
// as a Lambda environment variable / Secrets Manager secret — never in the repo
// — and Bedrock (if enabled) uses the Lambda role's IAM permissions.
package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	NewShipCheckStack(app, "ShipCheckStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
			Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
		},
	})
	app.Synth(nil)
}

// NewShipCheckStack defines the ShipCheck serverless stack.
func NewShipCheckStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	// --- DynamoDB: single table keyed by launch id, GSI on shareToken ---
	table := awsdynamodb.NewTable(stack, jsii.String("Launches"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("pk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST, // within the always-free allowance at low traffic
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,            // demo stack; destroy with the stack
	})
	table.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("share-index"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("shareToken"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		ProjectionType: awsdynamodb.ProjectionType_ALL,
	})

	// --- Lambda: pre-built Go binary on provided.al2023 (arm64) ---
	// The binary is built by scripts/build-lambda.sh into infra/lambda-build/
	// as "bootstrap" before `cdk deploy`.
	fn := awslambda.NewFunction(stack, jsii.String("Api"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(jsii.String("lambda-build"), nil),
		MemorySize:   jsii.Number(256),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Environment: &map[string]*string{
			"SHIPCHECK_TABLE":       table.TableName(),
			"SHIPCHECK_SHARE_INDEX": jsii.String("share-index"),
			// Default advisor is the free deterministic fallback — no key.
			"SHIPCHECK_MODEL_PROVIDER": jsii.String("fallback"),
		},
	})

	// Least-privilege: the Lambda may read/write the launches table and query
	// the GSI; nothing else. (Bedrock invoke is granted only when the Bedrock
	// provider is enabled — omitted here since the default is the fallback.)
	table.GrantReadWriteData(fn)

	// --- API Gateway HTTP API routing /api/* to the Lambda ---
	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("HttpApi"), &awsapigatewayv2.HttpApiProps{
		CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
			AllowOrigins: jsii.Strings("*"),
			AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{awsapigatewayv2.CorsHttpMethod_ANY},
			AllowHeaders: jsii.Strings("*"),
		},
	})
	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("ApiIntegration"), fn, nil)
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/api/{proxy+}"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_ANY},
		Integration: integration,
	})

	// The HTTP API endpoint host (strip the https:// scheme for the origin).
	apiDomain := awscdk.Fn_Select(jsii.Number(2),
		awscdk.Fn_Split(jsii.String("/"), httpApi.ApiEndpoint(), jsii.Number(3)))

	// --- S3 + CloudFront serving the SPA (HTTPS by default) ---
	siteBucket := awss3.NewBucket(stack, jsii.String("SpaBucket"), &awss3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(), // served only via CloudFront OAC
	})

	// API origin: the API Gateway HTTP API, reached over HTTPS.
	apiOrigin := awscloudfrontorigins.NewHttpOrigin(apiDomain, &awscloudfrontorigins.HttpOriginProps{
		ProtocolPolicy: awscloudfront.OriginProtocolPolicy_HTTPS_ONLY,
	})

	distribution := awscloudfront.NewDistribution(stack, jsii.String("Spa"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin:               awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(siteBucket, nil),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
		},
		// Route /api/* to the API Gateway origin on the SAME CloudFront domain,
		// so the SPA's relative /api/* calls work with no CORS and no separate
		// endpoint. The API is dynamic: disable caching and forward everything.
		AdditionalBehaviors: &map[string]*awscloudfront.BehaviorOptions{
			"/api/*": {
				Origin:               apiOrigin,
				ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
				AllowedMethods:       awscloudfront.AllowedMethods_ALLOW_ALL(),
				CachePolicy:          awscloudfront.CachePolicy_CACHING_DISABLED(),
				// Forward all viewer headers (except Host, which must be the
				// origin's) plus the query string and body to the API.
				OriginRequestPolicy: awscloudfront.OriginRequestPolicy_ALL_VIEWER_EXCEPT_HOST_HEADER(),
			},
		},
		DefaultRootObject: jsii.String("index.html"),
		// No SPA error-fallback is needed: the app uses HASH routing, so the
		// browser only ever requests "/" and static assets from the origin —
		// every in-app route lives in the "#/..." fragment the server never
		// sees. A distribution-wide 403/404 -> index.html rewrite would also
		// mask legitimate API 404s (e.g. a disabled share token), so it is
		// deliberately omitted.
	})

	// Deploy the built SPA assets (web/dist) to the bucket.
	awss3deployment.NewBucketDeployment(stack, jsii.String("SpaDeploy"), &awss3deployment.BucketDeploymentProps{
		Sources:           &[]awss3deployment.ISource{awss3deployment.Source_Asset(jsii.String("../web/dist"), nil)},
		DestinationBucket: siteBucket,
		Distribution:      distribution,
		DistributionPaths: jsii.Strings("/*"),
	})

	// Outputs: the CloudFront URL (SPA) and the API endpoint.
	awscdk.NewCfnOutput(stack, jsii.String("SpaUrl"), &awscdk.CfnOutputProps{
		Value: jsii.String("https://" + *distribution.DistributionDomainName()),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ApiUrl"), &awscdk.CfnOutputProps{
		Value: httpApi.ApiEndpoint(),
	})

	return stack
}
