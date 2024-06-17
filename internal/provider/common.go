package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	dbsTargetModel "github.com/joescharf/dbsnapper/v2/models/target"
)

func TFToResourceModel(ctx context.Context, tf *TargetResourceModel) (*TargetResourceModel, error) {

	// Initialize pointer fields
	if tf.Snapshot == nil {
		tf.Snapshot = new(targetSnapshotModel)
	}
	if tf.Share == nil {
		tf.Share = new(targetShareModel)
	}
	if tf.Sanitize == nil {
		tf.Sanitize = new(targetSanitizeModel)
	}

	// Read the list of SSOGroups from the Share attribute
	if !tf.Share.SSOGroups.IsNull() {
		elements := make([]types.String, 0, len(tf.Share.SSOGroups.Elements()))
		diags := tf.Share.SSOGroups.ElementsAs(ctx, &elements, false)
		if diags.HasError() {
			return tf, fmt.Errorf("Error reading SSOGroups: %s", diags)
		}
		ctx = tflog.SetField(ctx, "ID", tf.ID.String())
		ctx = tflog.SetField(ctx, "elements", elements)
		tflog.Debug(ctx, "TFToResourceModel -  TargetResourceModel")

	}
	return tf, nil
}

func ResourceModelToAPIRequest(ctx context.Context, resourceModel *TargetResourceModel) (*dbsTargetModel.Target, error) {
	targetRequest := new(dbsTargetModel.Target)
	targetRequest.Sanitize = dbsTargetModel.SanitizeCfg{}
	targetRequest.Share = dbsTargetModel.ShareCfg{}

	// Copy Snapshot fields
	if resourceModel.Snapshot != nil {
		targetRequest.Snapshot.SrcURL = resourceModel.Snapshot.SrcURL.ValueString()
		targetRequest.Snapshot.DstURL = resourceModel.Snapshot.DstURL.ValueString()
		targetRequest.Snapshot.SrcBytes = resourceModel.Snapshot.SrcBytes.ValueInt64()
	}

	// Copy the optional Sanitize fields
	if resourceModel.Sanitize != nil {
		targetRequest.Sanitize.DstURL = resourceModel.Sanitize.DstURL.ValueString()
		targetRequest.Sanitize.Query = resourceModel.Sanitize.Query.ValueString()
	}

	// Copy the optional Share fields
	if resourceModel.Share != nil {
		// Read from the plan using ElementsAs
		elements := make([]types.String, 0, len(resourceModel.Share.SSOGroups.Elements()))
		diag := resourceModel.Share.SSOGroups.ElementsAs(ctx, &elements, false)
		if diag.HasError() {
			return targetRequest, fmt.Errorf("Error reading SSOGroups: %s", diag)
		}
		// Copy the elements into the API request
		targetRequest.Share.SsoGroups = make([]string, len(elements))
		for i, v := range elements {
			targetRequest.Share.SsoGroups[i] = v.ValueString()
		}
	}

	uid, _ := uuid.Parse(resourceModel.ID.ValueString())

	targetRequest.ID = uid
	targetRequest.Name = resourceModel.Name.ValueString()
	targetRequest.Status = resourceModel.Status.ValueString()
	targetRequest.Messages = resourceModel.Messages.ValueString()

	ctx = tflog.SetField(ctx, "ID", targetRequest.ID.String())
	tflog.Debug(ctx, "PlanToApiRequest - targetRequest")

	return targetRequest, nil
}

func APIResponseToResourceModel(ctx context.Context, targetApiResponse *dbsTargetModel.Target, resourceModel *TargetResourceModel) (*TargetResourceModel, error) {
	resourceModel.ID = types.StringValue(targetApiResponse.ID.String())
	resourceModel.Name = types.StringValue(targetApiResponse.Name)
	resourceModel.Status = types.StringValue(targetApiResponse.Status)
	resourceModel.Messages = types.StringValue(targetApiResponse.Messages)

	// Snapshot
	resourceModel.Snapshot.SrcURL = types.StringValue(targetApiResponse.Snapshot.SrcURL)
	resourceModel.Snapshot.DstURL = types.StringValue(targetApiResponse.Snapshot.DstURL)
	resourceModel.Snapshot.SrcBytes = types.Int64Value(targetApiResponse.Snapshot.SrcBytes)

	// Sanitize
	if targetApiResponse.Sanitize != (dbsTargetModel.SanitizeCfg{}) {
		resourceModel.Sanitize.DstURL = types.StringValue(targetApiResponse.Sanitize.DstURL)
		resourceModel.Sanitize.Query = types.StringValue(targetApiResponse.Sanitize.Query)
	}
	// Share
	if targetApiResponse.Share.SsoGroups != nil {
		l, diag := types.ListValueFrom(ctx, types.StringType, targetApiResponse.Share.SsoGroups)
		if diag.HasError() {
			return resourceModel, fmt.Errorf("Error mapping SSOGroups: %s", diag)
		}
		resourceModel.Share.SSOGroups = l
	}
	resourceModel.CreatedAt = types.StringValue(targetApiResponse.CreatedAt)
	resourceModel.UpdatedAt = types.StringValue(targetApiResponse.UpdatedAt)

	return resourceModel, nil
}
