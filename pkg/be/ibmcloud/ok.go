package ibmcloud

import (
	"context"

	"github.com/IBM/vpc-go-sdk/vpcv1"
)

// Validate that our vpc service works
// TODO: this should accept no arguments and be a method on an instance that we return
func (backend Backend) Ok(ctx context.Context, initOk bool) error {
	limit := int64(1)
	resourceGroupId := backend.config.ResourceGroup.GUID

	_, _, err := backend.vpcService.ListVpcs(&vpcv1.ListVpcsOptions{
		Limit:           &limit,
		ResourceGroupID: &resourceGroupId,
	})

	return err
}
