package cloudflare

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/cloudflare/datacloudflarezones"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/cloudflare/record"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeforwardingrule"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeglobaladdress"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslcertificate"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computetargethttpsproxy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/loadbalancer"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
}

type Config struct {
	Project project.Project

	// SharedSecretsProjectID to source Cloudflare configuration
	SharedSecretsProjectID string

	Spec spec.EnvironmentDomainCloudflareSpec

	// Target LB setup for Cloudflare to route requests to
	Target loadbalancer.Output
}

// New sets up an external Cloudflare frontend for a load balancer target:
//
//	Cloudflare WAF -> ExternalAddress -> ForwardingRule -> HTTPSProxy -> Target
//
// This is partly based on the infrastructure generated by the Cloud Run Integration
// Custom Domains - Google Cloud Load Balancing and this old blog post:
// https://cloud.google.com/blog/topics/developers-practitioners/serverless-load-balancing-terraform-hard-way
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	// Create an SSL certificate from a secret in the shared secrets project
	sslCert := computesslcertificate.NewComputeSslCertificate(scope,
		id.ResourceID("origin-cert"),
		&computesslcertificate.ComputeSslCertificateConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: config.Project.ProjectId(),

			PrivateKey: &gsmsecret.Get(scope, id.SubID("secret-origin-private-key"), gsmsecret.DataConfig{
				Secret:    "SOURCEGRAPH_WILDCARD_KEY",
				ProjectID: config.SharedSecretsProjectID,
			}).Value,
			Certificate: &gsmsecret.Get(scope, id.SubID("secret-origin-cert"), gsmsecret.DataConfig{
				Secret:    "SOURCEGRAPH_WILDCARD_CERT",
				ProjectID: config.SharedSecretsProjectID,
			}).Value,

			Count: pointers.Float64(1),

			Lifecycle: &cdktf.TerraformResourceLifecycle{
				CreateBeforeDestroy: pointers.Ptr(true),
			},
		})

	// Set up an HTTPS proxy to route incoming HTTPS requests to our target's
	// URL map, which handles load balancing for a service.
	targetHTTPSProxy := computetargethttpsproxy.NewComputeTargetHttpsProxy(scope,
		id.ResourceID("https-proxy"),
		&computetargethttpsproxy.ComputeTargetHttpsProxyConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: config.Project.ProjectId(),
			UrlMap:  config.Target.URLMap.Id(),
			SslCertificates: pointers.Ptr([]*string{
				sslCert.Id(),
			}),
			SslPolicy: computesslpolicy.NewComputeSslPolicy(
				scope,
				id.ResourceID("ssl-policy"),
				&computesslpolicy.ComputeSslPolicyConfig{
					Name:    pointers.Ptr(id.DisplayName()),
					Project: config.Project.ProjectId(),

					Profile:       pointers.Ptr("MODERN"),
					MinTlsVersion: pointers.Ptr("TLS_1_2"),
				},
			).Id(),
		})

	// Set up an external address to receive traffic
	externalAddress := computeglobaladdress.NewComputeGlobalAddress(
		scope,
		id.ResourceID("external-address"),
		&computeglobaladdress.ComputeGlobalAddressConfig{
			Name:        pointers.Ptr(id.DisplayName()),
			Project:     config.Project.ProjectId(),
			AddressType: pointers.Ptr("EXTERNAL"),
			IpVersion:   pointers.Ptr("IPV4"),
		},
	)

	// Get the Cloudflare zone requested in configuration, and create a Cloudflare
	// record that points to our external address
	cfZone := datacloudflarezones.NewDataCloudflareZones(scope,
		id.ResourceID("domain"),
		&datacloudflarezones.DataCloudflareZonesConfig{
			Filter: &datacloudflarezones.DataCloudflareZonesFilter{
				Name: pointers.Ptr(config.Spec.Zone),
			},
		})
	_ = record.NewRecord(scope,
		id.ResourceID("record"),
		&record.RecordConfig{
			ZoneId: cfZone.Zones().Get(pointers.Float64(0)).Id(),
			Name:   &config.Spec.Subdomain,
			Type:   pointers.Ptr("A"),
			Value:  externalAddress.Address(),
			// Enable proxying to get WAF rules
			Proxied: pointers.Ptr(true),
		})

	// Forward traffic from the external address to the HTTPS proxy that then
	// routes request to our target
	_ = computeforwardingrule.NewComputeForwardingRule(scope,
		id.ResourceID("forwarding-rule"),
		&computeforwardingrule.ComputeForwardingRuleConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: config.Project.ProjectId(),

			IpAddress: externalAddress.Address(),
			PortRange: pointers.Ptr("443"),

			Target:              targetHTTPSProxy.Id(),
			LoadBalancingScheme: pointers.Ptr("EXTERNAL"),
		})

	return &Output{}, nil
}
