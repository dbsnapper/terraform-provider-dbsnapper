package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joescharf/dbsnapper/v2/storage"
)

func TFToSPResourceModel(ctx context.Context, tf *StorageProfileResourceModel) (*StorageProfileResourceModel, error) {

	// There are no sub-structs to initialize, so we can return the input

	return tf, nil
}

func SPResourceModelToAPIRequest(ctx context.Context, resourceModel *StorageProfileResourceModel) (*storage.StorageProfile, error) {
	spRequest := new(storage.StorageProfile)

	uid, _ := uuid.Parse(resourceModel.ID.ValueString())
	spRequest.ID = uid
	spRequest.Provider = resourceModel.Provider.ValueString()
	spRequest.Name = resourceModel.Name.ValueString()
	spRequest.Region = resourceModel.Region.ValueString()
	spRequest.AccountID = resourceModel.AccountID.ValueString()
	spRequest.AccessKey = resourceModel.AccessKey.ValueString()
	spRequest.SecretKey = resourceModel.SecretKey.ValueString()
	spRequest.Bucket = resourceModel.Bucket.ValueString()
	spRequest.Prefix = resourceModel.Prefix.ValueString()
	spRequest.Status = resourceModel.Status.ValueString()

	ctx = tflog.SetField(ctx, "ID", spRequest.ID.String())
	tflog.Debug(ctx, "PlanToApiRequest - storageProfileResource")

	return spRequest, nil
}

func APIResponseToSPResourceModel(ctx context.Context, spApiResponse *storage.StorageProfile, resourceModel *StorageProfileResourceModel) (*StorageProfileResourceModel, error) {
	resourceModel.ID = types.StringValue(spApiResponse.ID.String())
	resourceModel.Name = types.StringValue(spApiResponse.Name)
	resourceModel.Provider = types.StringValue(spApiResponse.Provider)
	resourceModel.Region = types.StringValue(spApiResponse.Region)
	resourceModel.AccountID = types.StringValue(spApiResponse.AccountID)
	resourceModel.AccessKey = types.StringValue(spApiResponse.AccessKey)
	resourceModel.SecretKey = types.StringValue(spApiResponse.SecretKey)
	resourceModel.Bucket = types.StringValue(spApiResponse.Bucket)
	resourceModel.Prefix = types.StringValue(spApiResponse.Prefix)
	resourceModel.Status = types.StringValue(spApiResponse.Status)
	resourceModel.CreatedAt = types.StringValue(spApiResponse.CreatedAt)
	resourceModel.UpdatedAt = types.StringValue(spApiResponse.UpdatedAt)

	return resourceModel, nil
}
