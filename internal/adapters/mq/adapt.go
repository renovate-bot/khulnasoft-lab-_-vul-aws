package mq

import (
	api "github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/mq/types"
	"github.com/khulnasoft-lab/defsec/pkg/providers/aws/mq"
	"github.com/khulnasoft-lab/defsec/pkg/state"
	defsecTypes "github.com/khulnasoft-lab/defsec/pkg/types"

	"github.com/khulnasoft-lab/vul-aws/internal/adapters"
	"github.com/khulnasoft-lab/vul-aws/pkg/concurrency"
)

type adapter struct {
	*adapters.RootAdapter
	api *api.Client
}

func init() {
	adapters.RegisterServiceAdapter(&adapter{})
}

func (a *adapter) Provider() string {
	return "aws"
}

func (a *adapter) Name() string {
	return "mq"
}

func (a *adapter) Adapt(root *adapters.RootAdapter, state *state.State) error {

	a.RootAdapter = root
	a.api = api.NewFromConfig(root.SessionConfig())
	var err error

	state.AWS.MQ.Brokers, err = a.getBrokers()
	if err != nil {
		return err
	}

	return nil
}

func (a *adapter) getBrokers() ([]mq.Broker, error) {

	a.Tracker().SetServiceLabel("Discovering brokers...")

	var apiBrokers []types.BrokerSummary
	var input api.ListBrokersInput
	for {
		output, err := a.api.ListBrokers(a.Context(), &input)
		if err != nil {
			return nil, err
		}
		apiBrokers = append(apiBrokers, output.BrokerSummaries...)
		a.Tracker().SetTotalResources(len(apiBrokers))
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	a.Tracker().SetServiceLabel("Adapting brokers...")
	return concurrency.Adapt(apiBrokers, a.RootAdapter, a.adaptBroker), nil
}

func (a *adapter) adaptBroker(apiBroker types.BrokerSummary) (*mq.Broker, error) {

	metadata := a.CreateMetadataFromARN(*apiBroker.BrokerArn)

	output, err := a.api.DescribeBroker(a.Context(), &api.DescribeBrokerInput{
		BrokerId: apiBroker.BrokerId,
	})
	if err != nil {
		return nil, err
	}

	return &mq.Broker{
		Metadata:     metadata,
		PublicAccess: defsecTypes.Bool(output.PubliclyAccessible, metadata),
		Logging: mq.Logging{
			Metadata: metadata,
			General:  defsecTypes.Bool(output.Logs != nil && output.Logs.General, metadata),
			Audit:    defsecTypes.Bool(output.Logs != nil && output.Logs.Audit, metadata),
		},
	}, nil
}
