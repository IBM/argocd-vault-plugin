package cmd

import (
	"bytes"
	"fmt"

	"github.com/IBM/argocd-vault-plugin/pkg/kube"
	"github.com/IBM/argocd-vault-plugin/pkg/kubernetesclient"
	"github.com/IBM/argocd-vault-plugin/pkg/vault"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewGenerateCommand initializes the generate command
func NewGenerateCommand() *cobra.Command {
	var configPath, secretName string

	var command = &cobra.Command{
		Use:   "generate <path>",
		Short: "Generate manifests from templates with Vault values",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("<path> argument required to generate manifests")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			files, err := listYamlFiles(path)
			if len(files) < 1 {
				return fmt.Errorf("no YAML files were found in %s", path)
			}
			if err != nil {
				return err
			}

			manifests, errs := readFilesAsManifests(files)
			if len(errs) != 0 {

				// TODO: handle multiple errors nicely
				return fmt.Errorf("could not read YAML files: %s", errs)
			}

			err = setConfig(secretName, configPath)
			if err != nil {
				return err
			}
			//
			vaultConfig, err := vault.NewConfig()
			if err != nil {
				return err
			}

			vaultClient := vaultConfig.Type

			err = vault.Login(vaultClient, vaultConfig)
			if err != nil {
				return err
			}

			for _, manifest := range manifests {

				template, err := kube.NewTemplate(manifest, vaultClient, vaultConfig.PathPrefix)
				if err != nil {
					return err
				}

				err = template.Replace()
				if err != nil {
					return err
				}

				output, err := template.ToYAML()
				if err != nil {
					return err
				}

				fmt.Printf("%s---\n", output)
			}

			return nil
		},
	}

	command.Flags().StringVarP(&configPath, "config-path", "c", "", "path to a file containing Vault configuration (YAML, JSON, envfile) to use")
	command.Flags().StringVarP(&secretName, "secret-name", "s", "", "name of a Kubernetes Secret containing Vault configuration data in the argocd namespace of your ArgoCD host (Only available when used in ArgoCD)")
	return command
}

func setConfig(secretName, configPath string) error {
	// If a secret name is passed, pull config from Kubernetes
	if secretName != "" {
		localClient, err := kubernetesclient.NewClient()
		if err != nil {
			return err
		}
		yaml, err := localClient.ReadSecret(secretName)
		if err != nil {
			return err
		}
		viper.SetConfigType("yaml")
		viper.ReadConfig(bytes.NewBuffer(yaml))
	}

	// If a config file path is passed, read in that file and overwrite all other
	if configPath != "" {
		viper.SetConfigFile(configPath)
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
	}

	return nil
}
