package explorer_test

import (
	"testing"

	pluginsconfig "github.com/ignite/cli/ignite/config/plugins"
	"github.com/ignite/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite/cli/ignite/services/plugin"
	envtest "github.com/ignite/cli/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGexExplorer(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		env     = envtest.New(t)
		app     = env.Scaffold("github.com/test/explorer")

		pluginRepo = "github.com/ignite/apps/explorer"

		assertPlugins = func(expectedLocalPlugins, expectedGlobalPlugins []pluginsconfig.Plugin) {
			localCfg, err := pluginsconfig.ParseDir(app.SourcePath())
			require.NoError(err)
			assert.ElementsMatch(expectedLocalPlugins, localCfg.Plugins, "unexpected local plugins")

			globalCfgPath, err := plugin.PluginsPath()
			require.NoError(err)
			globalCfg, err := pluginsconfig.ParseDir(globalCfgPath)
			require.NoError(err)
			assert.ElementsMatch(expectedGlobalPlugins, globalCfg.Plugins, "unexpected global plugins")
		}
	)

	env.Must(env.Exec("add explorer plugin",
		step.NewSteps(step.New(
			step.Exec(envtest.IgniteApp, "plugin", "add", pluginRepo),
			step.Workdir(app.SourcePath()),
		)),
	))

	// one local plugin expected
	assertPlugins(
		[]pluginsconfig.Plugin{
			{
				Path: pluginRepo,
			},
		},
		nil,
	)

	env.Must(env.Exec("run gex explorer help",
		step.NewSteps(step.New(
			step.Exec(
				envtest.IgniteApp,
				"e",
				"gex",
				"--help",
			),
			step.Workdir(app.SourcePath()),
		)),
	))
}
