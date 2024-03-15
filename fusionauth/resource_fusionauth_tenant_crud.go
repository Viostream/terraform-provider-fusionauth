package fusionauth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func createTenant(_ context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	client := i.(Client)
	tenant, diags := buildTenant(data)
	if diags != nil {
		return diags
	}

	t := fusionauth.TenantRequest{
		Tenant:         tenant,
		SourceTenantId: data.Get("source_tenant_id").(string),
	}
	client.FAClient.TenantId = ""

	var tid string
	if t, ok := data.GetOk("tenant_id"); ok {
		tid = t.(string)
	}
	resp, faErrs, err := client.FAClient.CreateTenant(tid, t)
	if err != nil {
		return diag.Errorf("CreateTenant err: %v", err)
	}

	if err := checkResponse(resp.StatusCode, faErrs); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(resp.Tenant.Id)
	return buildResourceDataFromTenant(resp.Tenant, data)
}

func readTenant(_ context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	client := i.(Client)
	id := data.Id()

	resp, faErrs, err := client.FAClient.RetrieveTenant(id)
	if err != nil {
		return diag.FromErr(err)
	}

	if resp.StatusCode == http.StatusNotFound {
		data.SetId("")
		return nil
	}
	if err := checkResponse(resp.StatusCode, faErrs); err != nil {
		return diag.FromErr(err)
	}

	return buildResourceDataFromTenant(resp.Tenant, data)
}

func updateTenant(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	client := i.(Client)
	tenant, diags := buildTenant(data)
	if diags != nil {
		return diags
	}

	t := map[string]interface{}{
		"Tenant":         tenant,
		"SourceTenantId": data.Get("source_tenant_id").(string),
	}

	tflog.Info(ctx, "Patching tenant with")

	// PatchTenant is used instead of UpdateTenant because we manage webhooks via the
	// webhook resource, and using the UpdateTenant method here deletes any webhooks
	// See https://github.com/gpsinsight/terraform-provider-fusionauth/issues/164
	resp, faErrs, err := client.FAClient.PatchTenant(data.Id(), t)
	if err != nil {
		return diag.Errorf("PatchTenant err: %v", err)
	}
	tflog.Info(ctx, fmt.Sprintf("Response: %+v", resp))
	if err := checkResponse(resp.StatusCode, faErrs); err != nil {
		return diag.FromErr(err)
	}

	return buildResourceDataFromTenant(resp.Tenant, data)
}

func deleteTenant(_ context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	client := i.(Client)
	resp, faErrs, err := client.FAClient.DeleteTenant(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := checkResponse(resp.StatusCode, faErrs); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
