// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// useStateForUnknownUnlessNull returns a plan modifier that copies the prior
// state value into the plan when state exists. Unlike UseStateForUnknown(),
// this handles the import case gracefully: when there is no prior state
// (e.g., import+apply in one operation), the value stays unknown rather than
// being forced to null, avoiding "inconsistent result after apply" errors.
func useStateForUnknownUnlessNull() planmodifier.String {
	return useStateForUnknownUnlessNullModifier{}
}

type useStateForUnknownUnlessNullModifier struct{}

func (m useStateForUnknownUnlessNullModifier) Description(_ context.Context) string {
	return "Copies the prior state value during plan when state exists, but allows unknown during import."
}

func (m useStateForUnknownUnlessNullModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateForUnknownUnlessNullModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If there's no prior state value, let the plan value stay as-is (unknown).
	// This handles import where there's no state yet.
	if req.StateValue.IsNull() {
		return
	}

	// If the plan value is unknown and we have a prior state value, use it.
	if resp.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}
