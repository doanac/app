package packager

import (
	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/docker/app/internal"
	"github.com/docker/app/internal/compose"
	"github.com/docker/app/types"
)

// ToCNAB creates a CNAB bundle from an app package
func ToCNAB(app *types.App, invocationImageName string) (*bundle.Bundle, error) {
	mapping := ExtractCNABParameterMapping(app.Parameters())
	flatParameters := app.Parameters().Flatten()
	parameters := map[string]bundle.ParameterDefinition{
		"docker.orchestrator": {
			DataType: "string",
			AllowedValues: []interface{}{
				"",
				"swarm",
				"kubernetes",
			},
			DefaultValue: "",
			Destination: &bundle.Location{
				EnvironmentVariable: "DOCKER_STACK_ORCHESTRATOR",
			},
			Metadata: bundle.ParameterMetadata{
				Description: "Orchestrator on which to deploy",
			},
		},
		"docker.kubernetes-namespace": {
			DataType: "string",
			Destination: &bundle.Location{
				EnvironmentVariable: "DOCKER_KUBERNETES_NAMESPACE",
			},
			Metadata: bundle.ParameterMetadata{
				Description: "Namespace in which to deploy",
			},
			DefaultValue: "",
		},
	}
	for name, envVar := range mapping.ParameterToCNABEnv {
		parameters[name] = bundle.ParameterDefinition{
			DataType: "string",
			Destination: &bundle.Location{
				EnvironmentVariable: envVar,
			},
			DefaultValue: flatParameters[name],
		}
	}
	var maintainers []bundle.Maintainer
	for _, m := range app.Metadata().Maintainers {
		maintainers = append(maintainers, bundle.Maintainer{
			Email: m.Email,
			Name:  m.Name,
		})
	}

	bundleImages, err := extractBundleImages(app.Composes())
	if err != nil {
		return nil, err
	}

	return &bundle.Bundle{
		Credentials: map[string]bundle.Location{
			"docker.context": {
				Path: "/cnab/app/context.dockercontext",
			},
		},
		Description: app.Metadata().Description,
		InvocationImages: []bundle.InvocationImage{
			{
				BaseImage: bundle.BaseImage{
					Image:     invocationImageName,
					ImageType: "docker",
				},
			},
		},
		Maintainers: maintainers,
		Name:        app.Metadata().Name,
		Version:     app.Metadata().Version,
		Parameters:  parameters,
		Actions: map[string]bundle.Action{
			internal.Namespace + "inspect": {
				Modifies: false,
			},
			internal.Namespace + "status": {
				Modifies: false,
			},
		},
		Images: bundleImages,
	}, nil
}

func extractBundleImages(composeFiles [][]byte) (map[string]bundle.Image, error) {
	_, images, err := compose.Load(composeFiles, func(v string) (string, error) { return v, nil })
	if err != nil {
		return nil, err
	}

	bundleImages := map[string]bundle.Image{}
	for serviceName, imageName := range images {
		bundleImages[serviceName] = bundle.Image{
			Description: imageName,
			BaseImage: bundle.BaseImage{
				Image:     imageName,
				ImageType: "docker",
			},
		}
	}
	return bundleImages, nil
}
