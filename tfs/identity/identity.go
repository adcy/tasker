package identity

import (
	"context"
	"errors"
	"tasker/ptr"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/identity"
	"github.com/spf13/viper"
)

func Get(ctx context.Context, conn *azuredevops.Connection) (string, error) {
	userFilter := viper.GetString("tfsUserFilter")

	client, err := identity.NewClient(ctx, conn)
	if err != nil {
		return "", err
	}

	identities, err := client.ReadIdentities(ctx, identity.ReadIdentitiesArgs{
		SearchFilter:    ptr.FromStr("General"),
		FilterValue:     ptr.FromStr(userFilter),
		QueryMembership: &identity.QueryMembershipValues.None,
	})
	if err != nil {
		return "", err
	}

	if identities == nil || len(*identities) == 0 {
		return "", errors.New("user identity not found")
	}

	if len(*identities) > 1 {
		return "", errors.New("user name not unique")
	}

	identity := (*identities)[0]

	return *identity.ProviderDisplayName, nil
}