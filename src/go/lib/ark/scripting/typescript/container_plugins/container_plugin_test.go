package container_plugins

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
)

/*
For the purpose of this test we will create Typescript VMs per test case
using the MustInitVM func so we can test plugins in isolation
*/
func TestLoad(t *testing.T) {
	type args struct {
		vm      *typescript.VirtualMachine
		plugins []workspace.Plugin
	}

	plugin := workspace.Plugin{
		Name:  "@ark/sre/microservice",
		Image: "gcr.io/[insert-google-project]/ark/plugins/microservice:latest",
	}

	tests := map[string]struct {
		args                   args
		tsFilePath             string
		wantErr                bool
		wantErrOnModResolution bool
	}{
		"single": {
			args: args{
				vm:      typescript.MustInitVM(nil),
				plugins: []workspace.Plugin{plugin},
			},
		},
		"invalid image": {
			args: args{
				vm: typescript.MustInitVM(nil),
				plugins: []workspace.Plugin{
					{
						Name:  "@ark/sre/microservice/",
						Image: "",
					}},
			},
			wantErrOnModResolution: true,
		},
		"import works": {
			args: args{
				vm:      typescript.MustInitVM(nil),
				plugins: []workspace.Plugin{plugin},
			},
			tsFilePath: "./testdata/simple_import.ts",
		},
		"executed imported module": {
			args: args{
				vm:      typescript.MustInitVM(nil),
				plugins: []workspace.Plugin{plugin},
			},
			tsFilePath: "./testdata/executes_fn_from_import.ts",
		},
		"got manifest back": {
			args: args{
				vm:      typescript.MustInitVM(nil),
				plugins: []workspace.Plugin{plugin},
			},
			tsFilePath: "./testdata/export.ts",
		},
		"got an error inside the plugin": {
			args: args{
				vm:      typescript.MustInitVM(nil),
				plugins: []workspace.Plugin{plugin},
			},
			tsFilePath:             "./testdata/error_within_a_plugin.ts",
			wantErrOnModResolution: true,
		},
		"cannot register the same plugin twice": {
			args: args{
				vm:      typescript.MustInitVM(nil),
				plugins: []workspace.Plugin{plugin, plugin},
			},
			wantErr: true,
		},
		"multiple modules": {
			args: args{
				vm: typescript.MustInitVM(nil),
				plugins: []workspace.Plugin{plugin,
					{
						Name:  "@ark/sre/vault-service-account",
						Image: "gcr.io/[insert-google-project]/ark/plugins/vault-k8s-sa:973721b27731109237e51e0dfe6891cdd2ce7dd0188a16b356b32b86c0fe6d32",
					}},
			},
			tsFilePath: "./testdata/multiples_plugin.ts",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := Load(context.Background(), tt.args.vm, tt.args.plugins); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			// execute the TS code if tsFilePath is present to perform assertions
			if tt.tsFilePath != "" {
				v, er := tt.args.vm.ResolveModule(tt.tsFilePath)
				if (er != nil) != tt.wantErrOnModResolution {
					t.Errorf("ResolveModule() error = %v", er)
				}

				mod := make(map[string]map[string]string)
				if err := tt.args.vm.Runtime.ExportTo(v, &mod); err != nil {
					t.Errorf("ExportT() error = %v", er)
				}

				exports := mod["exports"]
				manifest, ok := exports["default"]
				if ok {
					require.True(t, len(manifest) >= 1)
				}
			}
		})
	}
}
