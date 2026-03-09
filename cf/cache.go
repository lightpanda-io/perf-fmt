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

func (CacheNoop) Invalidate(context.Context, string) error {
	return nil
}

type CloudFrontCache struct {
	cf             *cloudfront.CloudFront
	distributionID string
}

func NewCloudFrontCache(sess *session.Session, distributionID string) *CloudFrontCache {
	return &CloudFrontCache{
		cf:             cloudfront.New(sess),
		distributionID: distributionID,
	}
}

func (c *CloudFrontCache) Invalidate(ctx context.Context, path string) error {
	paths := []string{path}

	_, err := c.cf.CreateInvalidationWithContext(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(c.distributionID),
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: aws.String(fmt.Sprintf("%d", time.Now().UnixNano())),
			Paths: &cloudfront.Paths{
				Quantity: aws.Int64(int64(len(paths))),
				Items:    aws.StringSlice(paths),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("cloudfront invalidation: %w", err)
	}

	return nil
}
