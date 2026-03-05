package cf

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

type Cache interface {
	Invalidate(ctx context.Context, path string) error
}

type CacheNoop struct{}

func (_ CacheNoop) Invalidate(_ context.Context, _ string) error {
	return nil
}

type CloudFrontCache struct {
	cf             *cloudfront.CloudFront
	distributionID string
}

func NewCloudFrontCache(distributionID string) (*CloudFrontCache, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("aws session: %w", err)
	}

	return &CloudFrontCache{
		cf:             cloudfront.New(sess),
		distributionID: distributionID,
	}, nil
}

func (c *CloudFrontCache) Invalidate(ctx context.Context, path string) error {
	paths := []*string{&path}

	_, err := c.cf.CreateInvalidationWithContext(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(c.distributionID),
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: aws.String(fmt.Sprintf("%d", time.Now().UnixNano())),
			Paths: &cloudfront.Paths{
				Quantity: aws.Int64(int64(len(paths))),
				Items:    paths,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("cloudfront invalidation: %w", err)
	}

	return nil
}
