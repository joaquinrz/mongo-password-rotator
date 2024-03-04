package keyvault

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joaquinrz/mongo-password-rotator/internal/config"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// UpdateSecret updates a secret in Azure Key Vault.
func UpdateSecret(cfg *config.Config) error {

	vaultURL := fmt.Sprintf("https://%s.vault.azure.net/", cfg.KeyVaultName)
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Printf("Could not authenticate to keyvault")
		return fmt.Errorf("failed to create Azure credential: %w", err)
	}

	client, err := azsecrets.NewClient(vaultURL, credential, nil)

	newPasswordBytes, err := os.ReadFile(cfg.NewPasswordFilePath)
	if err != nil {
		return fmt.Errorf("failed to read new password file: %w", err)
	}

	newPassword := string(newPasswordBytes)

	params := azsecrets.SetSecretParameters{
		Value: &newPassword,
	}

	_, err = client.SetSecret(context.Background(), cfg.KeyVaultCurrentSecretName, params, nil)
	if err != nil {
		return fmt.Errorf("failed to update secret in Azure Key Vault: %w", err)
	}

	return nil
}
