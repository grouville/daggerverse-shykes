

package main

import (
	"context"
)

type Bitgantter struct{}

func (m *Bitgantter) Build(
	// Application source code
	source *Directory,
) *Container {
	return source.DockerBuild()
}

func (m *Bitgantter) Test(ctx context.Context, app *Container) (string, error) {
	env, err := m.TestEnv(ctx, app, nil, nil, "app", "dev")
	if err != nil {
		return "", err
	}
	return env.WithExec([]string{"docker-compose", "exec", "app"}).Stdout(ctx)
}

// func (m *BitGantter) Test(
// 	ctx context.Context,
// 	// The application container to test
// 	app *Container,
// 	// A docker-compose file configuring the test environment
// 	// +optional
// 	config *File,
// 	// Pass env variables to docker-compose with an env-file
// 	// +optional
// 	envFile *File,
// 	// The image name of the compose service that will run the app
// 	// +default="app"
// 	// +optional
// 	name string,
// 	// The image tag of the compose service that will run the app
// 	// +default="dev"
// 	// +optional
// 	tag string,
// ) (string, error) {
// 	env, err := m.TestEnv(ctx, app, config, envFile, name, tag)
// 	if err != nil {
// 		return "", err
// 	}
// 	return env.
// 		WithExec([]string{
// 			"docker-compose", "exec", "test",
// 		}).
// 		Stdout(ctx)
// }

// A containerized test environment
func (m *Bitgantter) TestEnv(
	ctx context.Context,
	// The application container to test
	app *Container,
	// A docker-compose file configuring the test environment
	// +optional
	config *File,
	// Pass env variables to docker-compose with an env-file
	// +optional
	envFile *File,
	// The image name of the compose service that will run the app
	// +default="app"
	// +optional
	name string,
	// The image tag of the compose service that will run the app
	// +default="dev"
	// +optional
	tag string,
	) (*Container, error) {
		if config == nil {
			config = dag.CurrentModule().Source().File("./data/docker-compose.yml")
		}
		// Run an ephemeral Docker engine
		// FIXME: optionally connect to an external engine
		dockerd := dag.Docker().Engine()
		// Initialize a docker client
		docker := dag.Docker().Cli(DockerCliOpts{Engine: dockerd})
		// Import the app image into the docker engine
		appImage, err := docker.
			Import(app). // Import the app container into docker
			Duplicate(name, tag). // Tag the app container
			Ref(ctx)
		if err != nil {
		 	return nil, err
		}
		ctr := docker.Container().
			WithEnvVariable("APP_IMAGE", appImage).
			WithFile("/app/test/docker-compose.yml", config).
			WithWorkdir("/app/test")
		if envFile != nil {
			ctr = ctr.WithFile("/app/test/.envrc", envFile)
		}
		return ctr, nil
	}
