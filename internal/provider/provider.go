// Copyright (c) ZStack.io, Inc.

/*
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/

package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-zstack-modules/zsphere-sdk-go/pkg/client"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ZSphereProvider{}
var _ provider.ProviderWithFunctions = &ZSphereProvider{}
var _ provider.ProviderWithEphemeralResources = &ZSphereProvider{}

// ScaffoldingProvider defines the provider implementation.
type ZSphereProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type ZSphereProviderModel struct {
	Host            types.String `tfsdk:"host"`
	Port            types.Int64  `tfsdk:"port"`
	AccountName     types.String `tfsdk:"account_name"`
	AccountPassword types.String `tfsdk:"account_password"`
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	AccessKeySecret types.String `tfsdk:"access_key_secret"`
}

func (p *ZSphereProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zsphere"
	resp.Version = p.version
}

func (p *ZSphereProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "ZSphere Cloud MN HOST ip address. May also be provided via ZSphere_HOST environment variable.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "ZSphere Cloud MN API port. May also be provided via ZSphere_PORT environment variable.",
				Optional:    true,
			},
			/*
				"session_id": schema.StringAttribute{
					Description: "ZSphere Cloud Session id.",
					Optional:    true,
				},
			*/
			"account_name": schema.StringAttribute{
				Description: "Username for ZSphere API. May also be provided via ZSphere_ACCOUN_TNAME environment variable. " +
					"Required if using Account authentication.  Only supports the platform administrator account (`admin`). " +
					"Mutually exclusive with `access_key_id` and `access_key_secret`. " +
					"Using `access_key_id` and `access_key_secret` is the recommended approach for authentication, as it provides more flexibility and security.",
				Optional: true,
			},
			"account_password": schema.StringAttribute{
				Description: "Password for ZSphere API. May also be provided via ZSphere_ACCOUNT_PASSWORD environment variable." +
					"Required if using Account authentication.  Only supports the platform administrator account (`admin`). " +
					"Mutually exclusive with `access_key_id` and `access_key_secret`. " +
					"Using `access_key_id` and `access_key_secret` is the recommended approach for authentication, as it provides more flexibility and security.",
				Optional:  true,
				Sensitive: true,
			},
			"access_key_id": schema.StringAttribute{
				Description: "AccessKey ID for ZSphere API. Create AccessKey ID from MN,  Operational Management->Access Control->AccessKey Management. May also be provided via ZSphere_ACCESS_KEY_ID environment variable." +
					" Required if using AccessKey authentication. Mutually exclusive with `account_name` and `account_password`.",
				Optional: true,
			},
			"access_key_secret": schema.StringAttribute{
				Description: "AccessKey Secret for ZSphere API. May also be provided via ZSphere_ACCESS_KEY_SECRET environment variable." +
					" Required if using AccessKey authentication. Mutually exclusive with `account_name` and `account_password`.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *ZSphereProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	tflog.Info(ctx, "Configuring ZSphere client")

	//Retrieve provider data from configuration
	var config ZSphereProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown ZSphere Cloud API Host",
			"The provider cannt create the ZSphere Cloud API client as an unknown configuration value for the ZSphere Cloud API host."+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ZSphere_HOST environment variable.",
		)
	}

	if config.Port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Unknown ZSphere Cloud API Port",
			"The provider cannt create the ZSphere Cloud API client as an unknown configuration value for the ZSphere Cloud API port."+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ZSphere_PORT environment variable.",
		)
	}

	if config.AccountName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_name"),
			"Unknown ZSphere API account Username",
			"Either target apply the source of the value first, set the value statically in the configuration, or use the ZSphere_ACCOUNT_NAME environment variable.",
		)
	}

	if config.AccountPassword.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_password"),
			"Unknown ZSphere API account password",
			"Either target apply the source of the value first, set the value statically in the configuration, or use the ZSphere_ACCOUNT_PASSWORD environment variable.",
		)
	}

	if config.AccessKeyId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_key_id"),
			"Unknown ZSphere  access_key_id",
			"Either target apply the source of the value first, set the value statically in the configuration, or use the ZSphere_ACCESS_KEY_ID environment variable.",
		)
	}

	if config.AccessKeySecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_key_secret"),
			"Unknown ZSphere accessKeySecret",
			"Either target apply the source of the value first, set the value statically in the configuration, or use the ZSphere_ACCESS_KEY_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	//Defaukt value to environment vairiable, but override
	//with Terraform configuration value if set.

	port := 8080
	//sessionId := ""

	host := os.Getenv("ZSPHERE_HOST")
	portstr := os.Getenv("ZSPHERE_PORT")
	account_name := os.Getenv("ZSPHERE_ACCOUNT_NAME")
	account_password := os.Getenv("ZSPHERE_ACCOUNT_PASSWORD")
	access_key_id := os.Getenv("ZSPHERE_ACCESS_KEY_ID")
	access_key_secret := os.Getenv("ZSPHERE_ACCESS_KEY_SECRET")

	if portstr != "" {
		if portInt, err := strconv.Atoi(portstr); err == nil {
			port = portInt
		}
	}

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Port.IsNull() {
		port = int(config.Port.ValueInt64())
	}

	if !config.AccountName.IsNull() {
		account_name = config.AccountName.ValueString()
	}

	if !config.AccountPassword.IsNull() {
		account_password = config.AccountPassword.ValueString()
	}

	if !config.AccessKeyId.IsNull() {
		access_key_id = config.AccessKeyId.ValueString()
	}

	if !config.AccessKeySecret.IsNull() {
		access_key_secret = config.AccessKeySecret.ValueString()
	}
	/*
		if !config.SessionId.IsNull() {
			sessionId = config.SessionId.ValueString()
		}
	*/
	// If any of the expected configuration are missing, return
	// errors with provider-sepecific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing ZSphere API Host",
			"The provider cannot create the ZSphere API client as there is a missing or empty value for the ZSphere API host. "+
				"Set the host value in the configuration or use the ZSphere_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	//session id is just used for marketplace-server. Don't expose to user!
	if (account_name == "" || account_password == "") && (access_key_id == "" || access_key_secret == "") {
		resp.Diagnostics.AddError(
			"Missing ZSphere Authorization",
			"The provider cannot create the ZSphere API client as there is no ZSphere authorization. \n"+
				"Please set at least one authorization method: account_name + account_password OR access_key_id + access_key_secret.\n\n"+
				"account_name value can be set in the configuration or use the ZSphere_ACCOUNT_NAME environment variable\n"+
				"account_password value in the configuration or use the ZSphere_ACCOUNT_PASSWORD environment variable\n"+
				"access_key_id value in the configuration or use the ZSphere_ACCESS_KEY_ID environment variable\n"+
				"access_key_secret value in the configuration or use the ZSphere_ACCESS_KEY_SECRET environment variable\n")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var cli *client.ZSClient

	ctx = tflog.SetField(ctx, "zsphere_host", host)
	ctx = tflog.SetField(ctx, "zsphere_port", port)

	if account_name != "" && account_password != "" {
		ctx = tflog.SetField(ctx, "ZSphere_accountName", account_name)
		ctx = tflog.SetField(ctx, "ZSphere_accountPassword", account_password)
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "ZSphere_accountPassword")

		tflog.Debug(ctx, "Creating ZSphere client with account")
		cli = client.NewZSClient(client.NewZSConfig(host, port, "zstack").LoginAccount(account_name, account_password).ReadOnly(false).Debug(true))
		_, err := cli.Login()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create ZSphere API Client",
				"An unexpected error occurred when creating the ZSphere API client. "+
					"It might be due to an incorrect account name and password being set"+
					"If the error is not clear, please contact the provider developers.\n\n"+
					"ZSphere Client Error: "+err.Error(),
			)
			return
		}
	} else if access_key_id != "" && access_key_secret != "" {
		ctx = tflog.SetField(ctx, "ZSphere_accessKeyId", access_key_id)
		ctx = tflog.SetField(ctx, "ZSphere_accessKeySecret", access_key_secret)
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "ZSphere_accessKeySecret")

		tflog.Debug(ctx, "Creating ZSphere client with access key")
		cli = client.NewZSClient(client.NewZSConfig(host, port, "zstack").AccessKey(access_key_id, access_key_secret).ReadOnly(false).Debug(true))
		// no authorization validation! this access key may be invalid！
	}
	resp.DataSourceData = cli
	resp.ResourceData = cli

	tflog.Info(ctx, "Configured ZSphere client", map[string]any{"success": true})

}

func (p *ZSphereProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ImageResource,
		InstanceResource,
	}
}

func (p *ZSphereProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		//NewExampleEphemeralResource,
	}
}

func (p *ZSphereProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		ZSphereClusterDataSource,
		ZSphereZoneDataSource,
		ZSphereHostsDataSource,
		ZSphereImageStorageDataSource,
		ZSphereImageDataSource,
		ZSpherevmsDataSource,
		ZSphereL3NetworkDataSource,
		ZSpherePrimaryStorageDataSource,
	}
}

func (p *ZSphereProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ZSphereProvider{
			version: version,
		}
	}
}
