package fe

import (
	"fmt"
	"math/rand"
	"os"

	"lunchpail.io/pkg/be/helm"
	"lunchpail.io/pkg/compilation"
	"lunchpail.io/pkg/fe/linker"
	"lunchpail.io/pkg/fe/parser"
	"lunchpail.io/pkg/fe/transformer"
	"lunchpail.io/pkg/ir/llir"
)

type CompileOptions struct {
	linker.ConfigureOptions
	DryRun         bool
	Watch          bool
	UseThisRunName string
}

// TODO move into RestoreOptions
func valuesFromShrinkwrap(templatePath string, opts compilation.Options) (llir.Options, error) {
	shrinkwrappedOptions, err := compilation.RestoreOptions(templatePath)
	if err != nil {
		return opts, err
	} else {
		if opts.Namespace == "" {
			opts.Namespace = shrinkwrappedOptions.Namespace
		}
		// TODO here... how do we determine that boolean values were unset?
		if opts.ImagePullSecret == "" {
			opts.ImagePullSecret = shrinkwrappedOptions.ImagePullSecret
		}

		// careful: `--set x=3 --set x=4` results in x having
		// value 4, so we need to place the shrinkwrapped
		// options first in the list
		opts.OverrideValues = append(shrinkwrappedOptions.OverrideValues, opts.OverrideValues...)
		opts.OverrideFileValues = append(shrinkwrappedOptions.OverrideFileValues, opts.OverrideFileValues...)

		if opts.Queue == "" {
			opts.Queue = shrinkwrappedOptions.Queue
		}
		// TODO here... how do we determine that boolean values were unset?
		if opts.HasGpuSupport == false {
			opts.HasGpuSupport = shrinkwrappedOptions.HasGpuSupport
		}
		if !opts.CreateNamespace {
			opts.CreateNamespace = shrinkwrappedOptions.CreateNamespace
		}
	}

	return opts, nil
}

func PrepareForRun(opts CompileOptions) (llir.LLIR, compilation.Options, error) {
	stageOpts := compilation.StageOptions{}
	stageOpts.Verbose = opts.Verbose
	compilationName, templatePath, _, err := compilation.Stage(stageOpts)
	if err != nil {
		return llir.LLIR{}, opts.CompilationOptions, err
	}

	if updatedOpts, err := valuesFromShrinkwrap(templatePath, opts.CompilationOptions); err != nil {
		return llir.LLIR{}, opts.CompilationOptions, err
	} else {
		opts.CompilationOptions = updatedOpts
	}

	namespace := opts.CompilationOptions.Namespace
	if namespace == "" {
		namespace = compilationName
	}

	runname := opts.UseThisRunName
	if runname == "" {
		if generatedRunname, err := linker.GenerateRunName(compilationName); err != nil {
			return llir.LLIR{}, opts.CompilationOptions, err
		} else {
			runname = generatedRunname
		}
	}

	internalS3Port := rand.Intn(65536) + 1
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Using internal S3 port %d\n", internalS3Port)
	}

	yamlValues, dashdashSetValues, dashdashSetFileValues, queueSpec, err := linker.Configure(compilationName, runname, namespace, templatePath, internalS3Port, opts.ConfigureOptions)
	if err != nil {
		return llir.LLIR{}, opts.CompilationOptions, err
	}

	if !opts.Verbose {
		defer os.RemoveAll(templatePath)
	}

	// we need to instantiate the application's templates first...
	if yaml, err := helm.Template(runname, namespace, templatePath, yamlValues, helm.TemplateOptions{OverrideValues: dashdashSetValues, OverrideFileValues: dashdashSetFileValues, Verbose: opts.Verbose}); err != nil {
		return llir.LLIR{}, opts.CompilationOptions, err
	} else if hlir, err := parser.Parse(yaml); err != nil {
		return llir.LLIR{}, opts.CompilationOptions, err
	} else if ir, err := transformer.Lower(compilationName, runname, hlir, queueSpec, opts.ConfigureOptions.CompilationOptions, opts.Verbose); err != nil {
		return llir.LLIR{}, opts.CompilationOptions, err
	} else {
		return ir, opts.CompilationOptions, nil
	}
}