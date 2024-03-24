package main

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	apigateway "github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	acm "github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	cloudfront "github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	origins "github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	dynamodb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	awslambdaeventsources "github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	route53 "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	route53targets "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	s3assets "github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	s3deploy "github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	lambda "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"os"
	"strings"
)

type SkranAppSsrStackProps struct {
	awscdk.StackProps
	stackDetails StackConfigs
}

type StackConfigs struct {
	HostedZoneName  string `field:"optional"`
	AssetsSubdomain string `field:"optional"`
	SiteSubdomain   string `field:"optional"`
}

func SkranAppSsrStack(scope constructs.Construct, id string, props *SkranAppSsrStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	var cloudfrontDistribution cloudfront.Distribution
	hostedZoneName := props.stackDetails.HostedZoneName
	assetsSubdomain := props.stackDetails.AssetsSubdomain

	// Creates Origin Access Identity (OAI) to only allow CloudFront to get content
	cloudfrontOAI := cloudfront.NewOriginAccessIdentity(stack, jsii.String("skran-app-ssr-cloudfront-oai"), &cloudfront.OriginAccessIdentityProps{})

	// Creates S3 Bucket to store our static site content
	siteBucket := s3.NewBucket(stack, jsii.String("skran-app-ssr-assets"), &s3.BucketProps{
		BucketName:        jsii.String("skran-app-ssr-assets"),
		BlockPublicAccess: s3.BlockPublicAccess_BLOCK_ALL(),
		PublicReadAccess:  jsii.Bool(false),
		Versioned:         jsii.Bool(true),
	})

	// Adds a policy to the S3 Bucket that allows the OAI to get objects
	siteBucket.GrantRead(cloudfrontOAI, nil)

	cloudfrontDefaultBehavior := &cloudfront.BehaviorOptions{
		// Sets the S3 Bucket as the origin and tells CloudFront to use the created OAI to access it
		Origin: origins.NewS3Origin(siteBucket, &origins.S3OriginProps{
			OriginId:             jsii.String("skran-app-ssr-origin"),
			OriginAccessIdentity: cloudfrontOAI,
		}),
		ViewerProtocolPolicy: cloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}

	cloudfrontErrorResponses := &[]*cloudfront.ErrorResponse{
		{
			HttpStatus:         jsii.Number(403),
			ResponseHttpStatus: jsii.Number(403),
			ResponsePagePath:   jsii.String("/error.html"),
		},
	}

	// If Route 53 Hosted Zone is set, update AWS Certificate Manager, Route 53, and CloudFront accordingly
	if strings.TrimSpace(hostedZoneName) != "" {
		var assetsDomain string

		// If a assetsSubdomain is not set
		if strings.TrimSpace(assetsSubdomain) == "" {
			assetsDomain = hostedZoneName
		} else {
			assetsDomain = fmt.Sprintf("%s.%s", assetsSubdomain, hostedZoneName)
		}

		// Searches Route 53 for existing zone using hosted zone name
		hostedZone := route53.HostedZone_FromLookup(stack, jsii.String("skran-app-ssr-hosted-zone"), &route53.HostedZoneProviderProps{
			DomainName:  jsii.String(hostedZoneName),
			PrivateZone: jsii.Bool(false),
		})

		awscdk.Annotations_Of(hostedZone).AddInfo(jsii.String("Route 53 Hosted Zone is set"))

		// Creates an SSL/TLS certificate
		certificate := acm.Certificate_FromCertificateArn(stack, jsii.String("skran-app-ssr-cert"), jsii.String("arn:aws:acm:us-east-1:078577008688:certificate/338f23b9-ede8-4359-8a8f-f905510efb2c"))

		// Creates a new CloudFront Distribution with a custom Route 53 domain and custom SSL/TLS Certificate
		cloudfrontDistribution = cloudfront.NewDistribution(stack, jsii.String("skran-app-ssr-assets-cloudfront"), &cloudfront.DistributionProps{
			DefaultRootObject: jsii.String("index.html"),
			DefaultBehavior:   cloudfrontDefaultBehavior,
			ErrorResponses:    cloudfrontErrorResponses,
			Certificate:       certificate,
			DomainNames: &[]*string{
				jsii.String(assetsDomain),
			},
		})

		// Creates Route 53 record to point to the CloudFront Distribution
		publicEndpoint := route53.NewARecord(stack, jsii.String("skran-app-ssr-assets-cloudfront-public"), &route53.ARecordProps{
			Zone:       hostedZone,
			RecordName: jsii.String(assetsSubdomain),
			Target:     route53.RecordTarget_FromAlias(route53targets.NewCloudFrontTarget(cloudfrontDistribution)),
		})

		// Outputs public Route 53 endpoint
		awscdk.NewCfnOutput(stack, jsii.String("skran-app-ssr-assets-public-endpoint"), &awscdk.CfnOutputProps{
			Value: publicEndpoint.DomainName(),
		})
	} else {
		awscdk.Annotations_Of(stack).AddInfo(jsii.String("Route 53 Hosted Zone is NOT set"))

		// Creates a new CloudFront Distribution
		cloudfrontDistribution = cloudfront.NewDistribution(stack, jsii.String("skran-app-ssr-cloudfront"), &cloudfront.DistributionProps{
			DefaultRootObject: jsii.String("index.html"),
			DefaultBehavior:   cloudfrontDefaultBehavior,
			ErrorResponses:    cloudfrontErrorResponses,
		})
	}

	// Copies site assets from a local path to the S3 Bucket
	s3deploy.NewBucketDeployment(stack, jsii.String("skran-app-ssr-assets-deployment"), &s3deploy.BucketDeploymentProps{
		DestinationBucket: siteBucket,
		Sources: &[]s3deploy.ISource{
			s3deploy.Source_Asset(jsii.String("./public"), &s3assets.AssetOptions{}),
		},
		Distribution: cloudfrontDistribution,
		DistributionPaths: &[]*string{
			jsii.String("/*"),
		},
	})

	templates := s3.NewBucket(stack, jsii.String("skran-app-ssr-templates"), &s3.BucketProps{
		BucketName:        jsii.String("skran-app-ssr-templates"),
		BlockPublicAccess: s3.BlockPublicAccess_BLOCK_ALL(),
		PublicReadAccess:  jsii.Bool(false),
		Versioned:         jsii.Bool(true),
	})

	// Copies site assets from a local path to the S3 Bucket
	s3deploy.NewBucketDeployment(stack, jsii.String("skran-app-ssr-templates-deployment"), &s3deploy.BucketDeploymentProps{
		DestinationBucket: templates,
		Sources: &[]s3deploy.ISource{
			s3deploy.Source_Asset(jsii.String("./templates"), &s3assets.AssetOptions{}),
		},
	})

	var siteDomain string

	if strings.TrimSpace(props.stackDetails.SiteSubdomain) == "" {
		siteDomain = hostedZoneName
	} else {
		siteDomain = fmt.Sprintf("%s.%s", props.stackDetails.SiteSubdomain, hostedZoneName)
	}

	hostedZone := route53.HostedZone_FromLookup(stack, jsii.String("skran-app"), &route53.HostedZoneProviderProps{
		DomainName:  jsii.String(hostedZoneName),
		PrivateZone: jsii.Bool(false),
	})

	siteCert := acm.NewCertificate(stack, jsii.String("skran-app-ssr-site-cert"), &acm.CertificateProps{
		DomainName: jsii.String(siteDomain),
		Validation: acm.CertificateValidation_FromDns(hostedZone),
	})

	// Outputs CloudFront endpoint
	awscdk.NewCfnOutput(stack, jsii.String("skran-app-ssr-assets-cloudfront-endpoint"), &awscdk.CfnOutputProps{
		Value: cloudfrontDistribution.DistributionDomainName(),
	})

	// Outputs S3 Bucket endpoint (to show that it's not public)
	awscdk.NewCfnOutput(stack, jsii.String("skran-app-ssr-assets-endpoint"), &awscdk.CfnOutputProps{
		Value: siteBucket.BucketDomainName(),
	})

	ssrHandler := lambda.NewGoFunction(stack, jsii.String("skran-app-ssr-home"), &lambda.GoFunctionProps{
		FunctionName: jsii.String("skran-app-ssr-home"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Architecture: awslambda.Architecture_ARM_64(),
		Entry:        jsii.String("./src"),
		Environment:  &map[string]*string{"TEMPLATES": templates.BucketName(), "ASSETS_DOMAIN": jsii.String("https://assets.skran.app"), "TEMPLATE_DIR": jsii.String("/tmp")},
		Bundling: &lambda.BundlingOptions{
			GoBuildFlags: jsii.Strings(`-ldflags "-s -w"`),
		},
	})

	ssr := apigateway.NewLambdaRestApi(stack, jsii.String("skran-ssr-app-rest"), &apigateway.LambdaRestApiProps{
		DomainName: &apigateway.DomainNameOptions{
			DomainName:  jsii.String(siteDomain),
			Certificate: siteCert,
		},
		Handler: ssrHandler,
	})

	trigger := lambda.NewGoFunction(stack, jsii.String("skran-ssr-app-trigger"), &lambda.GoFunctionProps{
		FunctionName: jsii.String("skran-app-ssr-trigger"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Architecture: awslambda.Architecture_ARM_64(),
		Entry:        jsii.String("./trigger"),
		Bundling: &lambda.BundlingOptions{
			GoBuildFlags: jsii.Strings(`-ldflags "-s -w"`),
		},
		RetryAttempts: aws.Float64(0),
	})

	table := dynamodb.NewTable(stack, jsii.String("skran-ssr-app-table"), &dynamodb.TableProps{
		PartitionKey: &dynamodb.Attribute{Name: jsii.String("Primary"), Type: dynamodb.AttributeType_STRING},
		SortKey:      &dynamodb.Attribute{Name: jsii.String("Sort"), Type: dynamodb.AttributeType_STRING},
		TableName:    jsii.String("SkranAppTable"),
		Stream:       dynamodb.StreamViewType_NEW_AND_OLD_IMAGES,
	})

	trigger.AddEventSource(awslambdaeventsources.NewDynamoEventSource(table, &awslambdaeventsources.DynamoEventSourceProps{
		StartingPosition: awslambda.StartingPosition_TRIM_HORIZON,
		RetryAttempts:    aws.Float64(0),
	}))

	table.GrantWriteData(trigger)
	table.GrantReadData(ssrHandler)
	templates.GrantRead(ssrHandler, "*")

	route53.NewARecord(stack, jsii.String("skran-app-ssr-route"), &route53.ARecordProps{
		Zone:       hostedZone,
		RecordName: jsii.String(siteDomain),
		Target:     route53.RecordTarget_FromAlias(route53targets.NewApiGateway(ssr)),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	SkranAppSsrStack(app, "SkranAppSsrStack", &SkranAppSsrStackProps{
		awscdk.StackProps{
			Env: env(),
		},
		StackConfigs{
			// Optional
			// Set to an existing public Route 53 Hosted Zone in your control e.g. "amazon.com". Otherwise, set to ""
			HostedZoneName: "skran.app",

			// Optional
			// Add a subdomain to the hosted zone e.g. "aws". Otherwise, set to ""
			AssetsSubdomain: "assets",
			SiteSubdomain:   "www",
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	//return &awscdk.Environment{
	//	Account: jsii.String("078577008688"),
	//	Region:  jsii.String("eu-west-1"),
	//}

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
