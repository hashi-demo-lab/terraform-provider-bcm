package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// nullIfRemovedFromConfig returns a plan modifier that sets the planned value
// to null when the attribute is removed from config. This is needed for
// Optional+Computed attributes where Terraform's default behavior is to use
// the prior state value when config is null (treating it as "computed").
// For attributes like management_network that should be clearable by removing
// them from config, this modifier ensures an Update is triggered.
func nullIfRemovedFromConfig() planmodifier.String {
	return nullIfRemovedFromConfigModifier{}
}

type nullIfRemovedFromConfigModifier struct{}

func (m nullIfRemovedFromConfigModifier) Description(_ context.Context) string {
	return "Sets the planned value to null when the attribute is removed from configuration."
}

func (m nullIfRemovedFromConfigModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m nullIfRemovedFromConfigModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the attribute is not in config (null), plan should be null.
	// This overrides Terraform's default behavior for Optional+Computed,
	// which would use the prior state value when config is null.
	if req.ConfigValue.IsNull() {
		resp.PlanValue = types.StringNull()
	}
}
