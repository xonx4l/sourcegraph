package client

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/awsbedrock"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/azureopenai"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/fireworks"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Get(
	logger log.Logger,
	events *telemetry.EventRecorder,
	endpoint string,
	provider conftypes.CompletionsProviderName,
	accessToken string,
) (types.CompletionsClient, error) {
	client, err := getBasic(endpoint, provider, accessToken)
	if err != nil {
		return nil, err
	}
	return newObservedClient(logger, events, client), nil
}

func getBasic(endpoint string, provider conftypes.CompletionsProviderName, accessToken string) (types.CompletionsClient, error) {
	switch provider {
	case conftypes.CompletionsProviderNameAnthropic:
		return anthropic.NewClient(httpcli.ExternalDoer, endpoint, accessToken), nil
	case conftypes.CompletionsProviderNameOpenAI:
		return openai.NewClient(httpcli.ExternalDoer, endpoint, accessToken), nil
	case conftypes.CompletionsProviderNameAzureOpenAI:
		azureAPIClient, err := azureopenai.GetAzureAPIClient(endpoint, accessToken)
		if err != nil {
			return nil, err
		}
		return azureopenai.NewClient(azureAPIClient)
	case conftypes.CompletionsProviderNameSourcegraph:
		return codygateway.NewClient(httpcli.ExternalDoer, endpoint, accessToken)
	case conftypes.CompletionsProviderNameFireworks:
		return fireworks.NewClient(httpcli.ExternalDoer, endpoint, accessToken), nil
	case conftypes.CompletionsProviderNameAWSBedrock:
		return awsbedrock.NewClient(httpcli.ExternalDoer, endpoint, accessToken), nil
	default:
		return nil, errors.Newf("unknown completion stream provider: %s", provider)
	}
}
