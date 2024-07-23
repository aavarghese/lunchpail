package init

import (
	"context"
	"golang.org/x/sync/errgroup"
	"lunchpail.io/pkg/lunchpail/images"
	"lunchpail.io/pkg/lunchpail/images/build"
)

type InitLocalOptions struct {
	BuildImages bool
	Verbose     bool
}

func Local(opts InitLocalOptions) error {
	errs, _ := errgroup.WithContext(context.Background())

	if err := getContainerCli(); err != nil {
		return err
	}

	errs.Go(func() error {
		return getKubectl()
	})

	errs.Go(func() error {
		if err := getKind(); err != nil {
			return err
		}
		return createKindCluster()
	})

	errs.Go(func() error {
		return getNvidia()
	})

	if err := errs.Wait(); err != nil {
		return err
	}

	if opts.BuildImages {
		bopts := build.BuildOptions{}
		bopts.Verbose = opts.Verbose
		if err := images.Build(bopts); err != nil {
			return err
		}
	}

	return nil
}
